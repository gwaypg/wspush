package wsnode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gwaylib/errors"
	"github.com/gwaylib/eweb/jsonp"
)

const (
	CALLBACK_PROTO_HTTP  = "http"
	CALLBACK_PROTO_QUIC  = "quic"
	CALLBACK_PROTO_HTTPS = "https"
)

var (
	ErrHashMatch = errors.New("hash not match")
	ErrNoConn    = errors.New("connetion not found")
)

type ErrCode struct {
	code  int
	msg   string
	debug string
}

func NewErrCode(code int, msg string) *ErrCode {
	return &ErrCode{code, msg, ""}
}

func (e *ErrCode) Code() int {
	return e.code
}
func (e *ErrCode) Msg() string {
	return e.msg
}
func (e *ErrCode) As(err error) *ErrCode {
	return &ErrCode{
		code:  e.code,
		msg:   e.msg,
		debug: err.Error(),
	}
}

type Proto struct {
	Sn    string       `json:"sn"`
	Uri   string       `json:"uri"`
	Param jsonp.Params `json:"param,omitempty"`
}

func NewReqProto(sn, uri string) *Proto {
	return &Proto{
		Sn:    sn,
		Uri:   uri,
		Param: make(jsonp.Params),
	}
}
func NewRespProto(req *Proto, code *ErrCode) *Proto {
	param := jsonp.Params{}
	param.Add("code", strconv.Itoa(code.code))
	param.Add("msg", code.msg)
	if len(code.debug) > 0 {
		param.Add("debug", code.debug)
	}
	return &Proto{
		Sn:    req.Sn,
		Uri:   req.Uri,
		Param: param,
	}
}

func (r *Proto) String() string {
	return fmt.Sprintf("%+v", *r)
}

func (r *Proto) Serial() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return data
}

func ParseProto(data []byte) (*Proto, error) {
	r := &Proto{}
	if err := json.Unmarshal(data, r); err != nil {
		return nil, errors.As(err, string(data))
	}
	return r, nil
}

type CallBack struct {
	Proto    string // https or quic
	URL      string
	Insecure bool
}
