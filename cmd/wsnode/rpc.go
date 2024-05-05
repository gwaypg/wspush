package main

import (
	"fmt"
	"strings"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/redis"
	"github.com/gwaypg/wspush/module/cache"
	"github.com/gwaypg/wspush/module/wsnode"
)

type serviceImpl struct {
	wsconn *WsConn
}

func NewService(wsconn *WsConn) wsnode.RpcService {
	return &serviceImpl{
		wsconn: wsconn,
	}
}

func (impl *serviceImpl) loadCallbackCache() error {
	// loading callback
	r := cache.GetRedis()
	result := [][]byte{}
	cursor := int64(0)
	for {
		nextCursor, scanResult, err := r.ScanKey(cursor, cache.REDIS_WSNODE_CALLBACK_PREFIX+"*", 10000)
		if err != nil {
			if redis.ErrNil == err {
				break
			}
			return errors.As(err)
		}
		cursor = nextCursor
		result = append(result, scanResult...)
		if len(result) >= 1000 {
			break
		}
	}
	if len(result) > 1000 {
		panic(fmt.Sprintf("unonly support %d>1000 callback node", cursor))
	}
	for _, key := range result {
		cb := &wsnode.CallBack{}
		if err := r.ScanJSON(string(key), cb); err != nil {
			return errors.As(err)
		}
		tags := strings.Split(string(key), cache.REDIS_WSNODE_CALLBACK_PREFIX)
		if err := impl.wsconn.SetCallback(tags[1], cb); err != nil {
			return errors.As(err, string(key))
		}
	}
	return nil
}

func (impl *serviceImpl) SetCallback(arg *wsnode.SetCallbackArg, ret *wsnode.SetCallbackRet) error {
	if err := impl.wsconn.SetCallback(arg.Tag, arg.Cb); err != nil {
		return errors.As(err)
	}
	// save to redis
	r := cache.GetRedis()
	if err := r.PutJSON(cache.REDIS_WSNODE_CALLBACK_PREFIX+arg.Tag, arg.Cb, 0); err != nil {
		log.Warn(errors.As(err))
	}
	return nil
}

func (impl *serviceImpl) Push(arg *wsnode.PushArg, ret *wsnode.PushRet) error {
	return impl.wsconn.Push(arg.Uid, arg.Data)
}

func (impl *serviceImpl) CreateTopic(arg *wsnode.CreateTopicArg, ret *wsnode.CreateTopicRet) error {
	return impl.wsconn.CreateTopic(arg.Uid, arg.Topic...)
}
func (impl *serviceImpl) DestoryTopic(arg *wsnode.DestoryTopicArg, ret *wsnode.DestoryTopicRet) error {
	return impl.wsconn.DestoryTopic(arg.Uid, arg.Topic...)
}
func (impl *serviceImpl) GetTopicOwner(arg *wsnode.GetTopicOwnerArg, ret *wsnode.GetTopicOwnerRet) error {
	owner, err := impl.wsconn.GetTopicOwner(arg.Topic)
	if err != nil {
		return errors.As(err)
	}
	ret.Uid = owner
	return nil
}
func (impl *serviceImpl) IsTopicMember(arg *wsnode.IsTopicMemberArg, ret *wsnode.IsTopicMemberRet) error {
	ret.IsMember = impl.wsconn.IsTopicMember(arg.Uid, arg.Topic)
	return nil
}

func (impl *serviceImpl) JoinTopic(arg *wsnode.JoinTopicArg, ret *wsnode.JoinTopicRet) error {
	return impl.wsconn.JoinTopic(arg.Uid, arg.Topic...)
}
func (impl *serviceImpl) LeaveTopic(arg *wsnode.LeaveTopicArg, ret *wsnode.LeaveTopicRet) error {
	return impl.wsconn.LeaveTopic(arg.Uid, arg.Topic...)
}

func (impl *serviceImpl) SendTopic(arg *wsnode.SendTopicArg, ret *wsnode.SendTopicRet) error {
	// TODO: gorutine pool?
	go func() {
		if err := impl.wsconn.SendTopic(arg.Topic, arg.Data); err != nil {
			log.Warn(errors.As(err))
		}
	}()
	return nil
}
