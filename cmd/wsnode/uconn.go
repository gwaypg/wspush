package main

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gwaylib/errors"
	"github.com/gwaypg/wspush/module/wsnode"
)

type UserConn struct {
	cid        string          // 客户端ID
	conn       *websocket.Conn // 当前连接
	timeout    time.Duration   // 存活的检查时间
	aliveTimer *time.Timer     // 存活检查定时器
	sendBuffer chan []byte     // 单发专用
	sendResult chan error      // 单发结果
	broadcast  chan []byte     // 广播专用
	exit       chan bool       // 退出事件
	exitEnd    chan bool       // 退出完成事件
}

func NewUserConn(connId string, conn *websocket.Conn, timeout time.Duration) *UserConn {
	u := &UserConn{
		cid:        connId,
		conn:       conn,
		timeout:    timeout,
		aliveTimer: time.NewTimer(timeout),
		sendBuffer: make(chan []byte, 1),
		sendResult: make(chan error, 1),
		broadcast:  make(chan []byte, 100),
		exit:       make(chan bool, 1),
		exitEnd:    make(chan bool, 1),
	}
	go u.DeamonSend()
	return u
}

func closeConn(conn *websocket.Conn, code int, msg string) error {
	if conn == nil {
		return nil
	}

	if err := conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, msg),
		time.Now().Add(time.Second),
	); err != nil {
		return errors.As(err)
	}
	return conn.Close()
}

func (u *UserConn) Close(code int, msg string) error {
	if u.conn == nil {
		return nil
	}

	u.exit <- true
	<-u.exitEnd
	close(u.sendBuffer)
	close(u.sendResult)
	close(u.broadcast)
	u.aliveTimer.Stop()
	closeConn(u.conn, code, msg)
	u.conn = nil // 置空
	return nil
}

func (u *UserConn) ResetAliveTime() {
	u.aliveTimer.Reset(u.timeout)
}

func (u *UserConn) DeamonSend() {
	for {
		select {
		case <-u.aliveTimer.C:
			// 检查连接是否还正常
			if err := u.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(30*time.Second)); err == nil {
				u.ResetAliveTime()
				continue
			} else {
				log.Debug(errors.As(err))
			}

			code := wsnode.CloseErrTimeout
			msg := "心跳包超时"
			// 未检测到存活事件, 清理用户
			go u.conn.CloseHandler()(code, msg)

			<-u.exit
			u.exitEnd <- true
			return
		case <-u.exit:
			u.exitEnd <- true
			return
		case p := <-u.sendBuffer:
			err := u.conn.WriteMessage(websocket.BinaryMessage, p)
			u.sendResult <- errors.As(err)
		case p := <-u.broadcast:
			log.Debugf("broadcast:%+v\n", string(p))
			err := u.conn.WriteMessage(websocket.BinaryMessage, p)
			if err != nil {
				log.Warn(errors.As(err))
			}
		}
	}
}

func (u *UserConn) Send(p []byte) error {
	if u.conn == nil {
		return errors.New("conn closed")
	}
	u.sendBuffer <- p
	err := <-u.sendResult
	return err
}

func (u *UserConn) Broadcast(p []byte) error {
	if u.conn == nil {
		return errors.New("conn closed")
	}
	u.broadcast <- p
	return nil
}

var (
	ErrCloseEvent = errors.New("close event")
)

func (u *UserConn) ReadMessage() ([]byte, error) {
	frameType, data, err := u.conn.ReadMessage()
	if err != nil {
		_, ok := err.(*websocket.CloseError)
		if ok {
			return nil, ErrCloseEvent.As(err)
		}
		return nil, errors.As(err, frameType, string(data))
	}
	if !(frameType == websocket.TextMessage || frameType == websocket.BinaryMessage) {
		return nil, errors.New("error frame type").As(frameType)
	}
	return data, nil
}

type UserMulConn struct {
	conns sync.Map // map[cid]UserConn
}

func (m *UserMulConn) Load(cid string) (*UserConn, bool) {
	conn, ok := m.conns.Load(cid)
	if !ok {
		return nil, false
	}
	return conn.(*UserConn), true
}
func (m *UserMulConn) Store(cid, conn interface{}) {
	m.conns.Store(cid, conn)
}
func (m *UserMulConn) Delete(cid interface{}) {
	m.conns.Delete(cid)
}
