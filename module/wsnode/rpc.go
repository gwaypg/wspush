package wsnode

import (
	"github.com/gwaylib/errors"
	"github.com/gwaylib/rpctry"
	"github.com/gwaypg/wspush/module/etc"
)

// Rpc exported method
type RpcService interface {
	// P2P send
	Push(arg *PushArg, ret *PushRet) error

	// create a topic
	CreateTopic(arg *CreateTopicArg, ret *CreateTopicRet) error
	// desctroy a topic
	DestoryTopic(arg *DestoryTopicArg, ret *DestoryTopicRet) error
	// get topic owner
	GetTopicOwner(arg *GetTopicOwnerArg, ret *GetTopicOwnerRet) error
	// is topic member
	IsTopicMember(arg *IsTopicMemberArg, ret *IsTopicMemberRet) error
	// join a topic
	JoinTopic(arg *JoinTopicArg, ret *JoinTopicRet) error
	// leave a topic
	LeaveTopic(arg *LeaveTopicArg, ret *LeaveTopicRet) error
	// send message to a topic
	SendTopic(arg *SendTopicArg, ret *SendTopicRet) error
}

var (
	RpcName   = etc.Etc.String("cmd/wsnode", "rpc_name")
	rpcClient = rpctry.NewClient(etc.Etc.String("cmd/wsnode", "rpc_client"))
)

type SetCallbackArg struct {
	Tag string
	Cb  *CallBack
}
type SetCallbackRet struct {
}

func SetCallback(arg *SetCallbackArg) (*SetCallbackRet, error) {
	ret := &SetCallbackRet{}
	if err := rpcClient.TryCall(RpcName+".SetCallback", arg, ret); err != nil {
		return ret, errors.As(err)
	}
	return ret, nil
}

type PushArg struct {
	Uid  string
	Data *Proto
}
type PushRet struct {
}

func Push(arg *PushArg) (*PushRet, error) {
	ret := &PushRet{}
	if err := rpcClient.TryCall(RpcName+".Push", arg, ret); err != nil {
		return ret, errors.As(err)
	}
	return ret, nil
}

type CreateTopicArg struct {
	Uid   string
	Topic []string
}
type CreateTopicRet struct {
}

func CreateTopic(arg *CreateTopicArg) (*CreateTopicRet, error) {
	ret := &CreateTopicRet{}
	if err := rpcClient.TryCall(RpcName+".CreateTopic", arg, ret); err != nil {
		return ret, errors.As(err)
	}
	return ret, nil
}

type DestoryTopicArg struct {
	Uid   string
	Topic []string
}
type DestoryTopicRet struct {
}

func DestoryTopic(arg *DestoryTopicArg) (*DestoryTopicRet, error) {
	ret := &DestoryTopicRet{}
	if err := rpcClient.TryCall(RpcName+".DestoryTopic", arg, ret); err != nil {
		return ret, errors.As(err)
	}
	return ret, nil
}

type GetTopicOwnerArg struct {
	Topic string
}
type GetTopicOwnerRet struct {
	Uid string
}

func GetTopicOwner(arg *GetTopicOwnerArg) (*GetTopicOwnerRet, error) {
	ret := &GetTopicOwnerRet{}
	if err := rpcClient.TryCall(RpcName+".GetTopicOwner", arg, ret); err != nil {
		return ret, errors.As(err)
	}
	return ret, nil
}

type JoinTopicArg struct {
	Uid   string
	Topic []string
}
type JoinTopicRet struct {
}

func JoinTopic(arg *JoinTopicArg) (*JoinTopicRet, error) {
	ret := &JoinTopicRet{}
	if err := rpcClient.TryCall(RpcName+".JoinTopic", arg, ret); err != nil {
		return ret, errors.As(err)
	}
	return ret, nil
}

type LeaveTopicArg struct {
	Uid   string
	Topic []string
}
type LeaveTopicRet struct {
}

func LeaveTopic(arg *LeaveTopicArg) (*LeaveTopicRet, error) {
	ret := &LeaveTopicRet{}
	if err := rpcClient.TryCall(RpcName+".LeaveTopic", arg, ret); err != nil {
		return ret, errors.As(err)
	}
	return ret, nil
}

type IsTopicMemberArg struct {
	Uid   string
	Topic string
}
type IsTopicMemberRet struct {
	IsMember bool
}

func IsTopicMember(arg *IsTopicMemberArg) (*IsTopicMemberRet, error) {
	ret := &IsTopicMemberRet{}
	if err := rpcClient.TryCall(RpcName+".IsTopicMember", arg, ret); err != nil {
		return ret, errors.As(err)
	}
	return ret, nil
}

type SendTopicArg struct {
	Topic string
	Data  *Proto
}
type SendTopicRet struct {
}

func SendTopic(arg *SendTopicArg) (*SendTopicRet, error) {
	ret := &SendTopicRet{}
	if err := rpcClient.TryCall(RpcName+".SendTopic", arg, ret); err != nil {
		return ret, errors.As(err)
	}
	return ret, nil
}
