// TODO: 在量大时可能会升级为连接集群
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gwaypg/wspush/module/wsnode"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/qlog"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/gwaylib/errors"
	"github.com/gwaylib/qsql"
	echo "github.com/labstack/echo/v4"
)

var (
	ErrNoConn = errors.New("no connect")
	ErrAuth   = errors.New("auth error")
	ErrArgs   = errors.New("err arguments")
)

type WsConn struct {
	userLk    sync.Mutex
	userConns sync.Map // map[uid]*UserMulConn

	roomLk sync.Mutex
	rooms  sync.Map // map[string]*Topic

	callbackLk sync.Mutex
	callback   sync.Map // map[string]string
}

func (ws *WsConn) LoadUser(uid string) (*UserMulConn, bool) {
	sConns, ok := ws.userConns.Load(uid)
	if !ok {
		return nil, false
	}
	return sConns.(*UserMulConn), true
}

func (ws *WsConn) LoadRoom(roomId string) (*Topic, bool) {
	room, ok := ws.rooms.Load(roomId)
	if !ok {
		return nil, false
	}
	return room.(*Topic), true
}
func (ws *WsConn) LoadCallback(tag string) (string, bool) {
	callback, ok := ws.callback.Load(tag)
	if !ok {
		return "", false
	}
	return callback.(string), true
}

func NewWsConn() *WsConn {
	return &WsConn{
		userConns: sync.Map{},
		rooms:     sync.Map{},
		callback:  sync.Map{},
	}
}

func (w *WsConn) SetCallback(tag string, cb *wsnode.CallBack) error {
	switch cb.Proto {
	case wsnode.CALLBACK_PROTO_QUIC, wsnode.CALLBACK_PROTO_HTTPS, wsnode.CALLBACK_PROTO_HTTP:
	default:
		return errors.New("unsupport proto").As(cb)
	}

	if len(tag) == 0 {
		return errors.New("tag not set")
	}
	if len(cb.URL) == 0 {
		return errors.New("URL not set")
	}

	w.callback.Store(tag, cb)
	return nil
}

func (w *WsConn) handleCb(cid, uid, token, tag string, req *wsnode.Proto) ([]byte, error) {
	cbI, ok := w.callback.Load(tag)
	if !ok {
		return nil, ErrArgs.As("callback not set").As(tag)
	}
	cb, _ := cbI.(*wsnode.CallBack)
	httpClient := &http.Client{}
	switch cb.Proto {
	case wsnode.CALLBACK_PROTO_QUIC:
		httpClient.Transport = &http3.RoundTripper{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cb.Insecure,
			},
			QUICConfig: &quic.Config{
				Tracer: qlog.DefaultTracer,
			},
		}
	case wsnode.CALLBACK_PROTO_HTTPS:
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cb.Insecure,
			},
		}
	default:
		// inore

	}
	// TODO: does open too many files?
	auth := url.Values{}
	auth.Add("cid", cid)
	auth.Add("uid", uid)
	auth.Add("token", token)
	resp, err := httpClient.Post(
		fmt.Sprintf("%s?%s", cb.URL, auth.Encode()), // for auth
		"application/json",
		bytes.NewReader(req.Serial()),
	)
	if err != nil {
		return nil, errors.As(err)
	}
	defer qsql.Close(resp.Body)
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.As(err)
	}
	switch resp.StatusCode {
	case 200:
		// pass
	case 401:
		return nil, ErrAuth.As(cb.URL, string(data), uid)
	default:
		return nil, errors.New(string(data)).As(cb.URL, resp.StatusCode, uid, token, *req)
	}
	return data, nil

}

// 守护读取并处理连接
func (w *WsConn) HandleConn(c echo.Context, conn *websocket.Conn) error {

	// 检查token
	cid := c.FormValue("cid") // client id
	uid := c.FormValue("uid")
	token := c.FormValue("token")
	tag := c.FormValue("tag")

	uConn := NewUserConn(cid, conn, 30*time.Second)
	// 登录回调
	if _, err := w.handleCb(
		cid, uid, token, tag,
		wsnode.NewReqProto(uuid.New().String(), "/user/login"),
	); err != nil {
		if ErrAuth.Equal(err) {
			//log.Info(errors.As(err))
			return uConn.Close(wsnode.CloseErrAuth, errors.As(err).Code())
		}
		if ErrArgs.Equal(err) {
			return uConn.Close(wsnode.CloseProtocolError, errors.As(err).Code())
		}
		log.Warn(errors.As(err))
		return uConn.Close(wsnode.CloseInternalServerErr, errors.As(err).Code())
	}

	// 检查是否已存在，若已存在，关闭之前的连接
	w.userLk.Lock()
	mulConn, ok := w.LoadUser(uid)
	if !ok {
		mulConn = &UserMulConn{}
		w.userConns.Store(uid, mulConn)
	}
	w.userLk.Unlock()

	mulConn.conns.Range(func(key, val interface{}) bool {
		conn := val.(*UserConn)
		if cid == conn.cid {
			// 关闭之前的连接
			log.Debugf("close history :%s, %s", uid, cid)
			if err := conn.Close(wsnode.CloseErrAuth, "新的登录取代了现有登录"); err != nil {
				log.Warn(errors.As(err))
			}
			mulConn.Delete(key)
		}
		return true
	})
	mulConn.Store(cid, uConn)

	defer func() {
		w.CloseWithConn(cid, uid, conn, websocket.CloseNormalClosure, "请求已完成")
	}()

	// 设置下线通知
	func(u *UserConn) {
		u.conn.SetCloseHandler(func(code int, text string) error {
			req := wsnode.NewReqProto(uuid.New().String(), "/user/logout")
			req.Param.AddAny("cid", u.cid)
			req.Param.AddAny("code", code)
			req.Param.AddAny("text", text)
			if _, err := w.handleCb(
				cid, uid, token, tag, req,
			); err != nil {
				log.Warn(errors.As(err))
			}

			return u.Close(code, text)
		})
	}(uConn) // 复制一份副本

	// 设置PING回应
	func(u *UserConn) {
		u.conn.SetPingHandler(func(message string) error {
			log.Debugf("ping handle:%s\n", message)
			// 读取到消息, 重置连接失效时间
			u.ResetAliveTime()
			if err := u.conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(5*time.Second)); err != nil {
				return nil
			} else if e, ok := err.(net.Error); ok && e.Temporary() {
				return nil
			}
			return nil
		})
	}(uConn)

	log.Debugf("logon:%s", uid)
	for {
		msgData, err := uConn.ReadMessage()
		if err != nil {
			if ErrCloseEvent.Equal(err) {
				return nil
			}
			log.Warn(errors.As(err))
			return uConn.Close(wsnode.CloseErrConn, "网络异常")
		}
		// 读取到消息, 重置连接失效时间
		uConn.ResetAliveTime()
		log.Debugf("recv req:%s", string(msgData))

		// decode data
		req := &wsnode.Proto{}
		if err := json.Unmarshal(msgData, req); err != nil {
			log.Warn(errors.As(err))
			return uConn.Close(wsnode.CloseProtocolError, "协议格式错误,请检查协议格式")
		}
		resp, err := w.handleCb(cid, uid, token, tag, req)
		if err != nil {
			if ErrAuth.Equal(err) {
				return uConn.Close(wsnode.CloseErrAuth, "鉴权失败")
			}
			log.Warn(errors.As(err))
			return uConn.Close(wsnode.CloseErrConn, "网络响应失败")
		}
		if err := uConn.Send(resp); err != nil {
			log.Warn(errors.As(err))
			return uConn.Close(wsnode.CloseErrConn, "网络响应失败")
		}
	}
}

// 主动关闭用户连接
func (w *WsConn) Close(cid, uid string, code int, msg string) error {
	mulConn, ok := w.LoadUser(uid)
	if !ok {
		return ErrNoConn.As(uid)
	}
	uConn, ok := mulConn.Load(cid)
	if !ok {
		return ErrNoConn.As(uid)
	}
	if err := uConn.Close(code, msg); err != nil {
		log.Warn(errors.As(err))
	}
	mulConn.Delete(cid)

	// clean user connection
	count := 0
	mulConn.conns.Range(func(key, val interface{}) bool {
		count++
		return false
	})
	if count == 0 {
		w.userConns.Delete(uid)
	}
	return nil
}

// 主动关闭用户连接，若找不到ID，再对conn进行关闭
func (w *WsConn) CloseWithConn(cid, uid string, conn *websocket.Conn, code int, msg string) {
	if err := w.Close(cid, uid, code, msg); err != nil {
		if !ErrNoConn.Equal(err) {
			log.Warn(errors.As(err))
			return
		}
		// close the default conn and ignore any error
		conn.Close()
		return
	}
}

func (w *WsConn) IsTopicMember(uid, topic string) bool {
	room, ok := w.LoadRoom(topic)
	if !ok {
		return false
	}

	// for owner
	if room.Owner == uid {
		return true
	}

	// for memeber
	_, ok = room.Member.Load(uid)
	return ok
}

// 用户是否在线
func (w *WsConn) IsOnline(uid, clientId string) bool {
	_, ok := w.userConns.Load(uid)
	return ok
}

// 实时推送, 若不可达，直接返回错误
func (w *WsConn) Push(uid string, p *wsnode.Proto) error {
	v, ok := w.userConns.Load(uid)
	if !ok {
		return ErrNoConn.As(uid, *p)
	}
	conns := v.([]*UserConn)

	wsData, err := json.Marshal(p)
	if err != nil {
		return errors.As(err)
	}
	result := make(chan error, len(conns))
	for _, conn := range conns {
		// TODO: 池化？
		go func() {
			result <- conn.Send(wsData)
		}()
	}
	var resultErr error
	for i := 0; i < len(conns); i++ {
		err := <-result
		if err == nil {
			// 确保至少有一个可达
			resultErr = nil
		} else {
			resultErr = err
			log.Info(errors.As(err))
		}
	}
	return errors.As(resultErr)
}

// 实时广播消息(这是不可靠的，如果必要，需要客户端发送确认收到协议)
// TODO: make history?
func (w *WsConn) SendTopic(topic string, p *wsnode.Proto) error {
	var room *Topic
	room, ok := w.LoadRoom(topic)
	if !ok {
		return errors.ErrNoData.As(topic)
	}

	wsData, err := json.Marshal(p)
	if err != nil {
		return errors.As(err)
	}

	room.Member.Range(func(uid, val interface{}) bool {
		mulConn, ok := w.LoadUser(uid.(string))
		if !ok {
			return true
		}

		mulConn.conns.Range(func(cid, val interface{}) bool {
			uConn := val.(*UserConn)
			if err := uConn.Broadcast(wsData); err != nil {
				// unexpect here
				log.Warn(errors.As(err))
			}
			return false
		})
		return true
	})
	return nil
}

// 创建某个频道
// 当前已知的主题:
// 系统主题：/all/system
// 场景主题：/all/scene/:id
// 世界主题：暂不开放，待进一步定义
func (w *WsConn) CreateTopic(uid string, topics ...string) error {
	for _, topic := range topics {
		room, ok := w.LoadRoom(topic)
		if ok {
			if room.Owner != uid {
				return errors.New("topic is existed").As(topic)
			}
			return nil
		}

		// create a new one
		topicRoom := &Topic{
			Owner:  uid,
			Member: sync.Map{},
		}
		w.rooms.Store(topic, topicRoom)
	}
	return nil
}

func (w *WsConn) DestoryTopic(uid string, topics ...string) error {
	for _, topic := range topics {
		room, ok := w.LoadRoom(topic)
		if !ok {
			return nil
		}
		if room.Owner != uid {
			return errors.New("topic is existed").As(topic)
		}
		w.rooms.Delete(topic)
	}
	return nil
}

func (w *WsConn) GetTopicOwner(topic string) (string, error) {
	room, ok := w.LoadRoom(topic)
	if !ok {
		return "", errors.ErrNoData.As(topic)
	}
	return room.Owner, nil
}

// 订阅某个已创建的频道
func (w *WsConn) JoinTopic(uid string, topics ...string) error {
	for _, topic := range topics {
		val, ok := w.rooms.Load(topic)
		if !ok {
			return errors.ErrNoData.As(topic)
		}
		topicRoom := val.(*Topic)
		topicRoom.Member.Store(uid, true)
		w.rooms.Store(topic, topicRoom)
	}
	return nil
}

// 取消某个频道的订阅
func (w *WsConn) LeaveTopic(uid string, topics ...string) error {
	for _, topic := range topics {
		val, ok := w.rooms.Load(topic)
		if !ok {
			return nil
		}
		topicRoom := val.(*Topic)
		topicRoom.Member.Delete(uid)
		w.rooms.Store(topic, topicRoom)
	}
	return nil
}
