package wsnode

import "github.com/gorilla/websocket"

const (
	// 状态码
	// 0 - 999 ：未使用。
	// 1000 - 2999 ：此范围的保留给本协议使用的，它以后的校订和扩展将以永久的和容易获取  公开规范里指定。
	// 3000 - 3999 ：此范围保留给类库、框架和应用程序使用。这些状态码直接在IANA注册。  本协议未定义如何解释这些状态码。
	// 4000 - 4999 ：此范围保留给私用，且不能注册。这些值可以在预先达成一致的WebSocket   应用程序间使用。本协议未定义如何解释这些值。

	// 1000：表示正常关闭，意味着连接建立的目的已完成。
	CloseNormalClosure = websocket.CloseNormalClosure

	// 系统定义
	// 1001：表示终端离开，例如服务器关闭或浏览器导航到其他页面。
	CloseGoingAway = websocket.CloseGoingAway
	// 1002：表示终端因为协议错误而关闭连接。
	CloseProtocolError = websocket.CloseProtocolError
	// 1003：表示终端因为接收到不能接受的数据而关闭（例如，只明白文本数据的终端可能发送这个，如果它接收到二进制消息）。
	CloseUnsupportedData = websocket.CloseUnsupportedData

	/////////////////////
	// 以下状态码需要重试(1004-4000)
	// //////////////////
	// 1005：保留。且终端必须不在控制帧里设置作为状态码。它是指定给应用程序而非作为状态码 使用的，用来指示没有状态码出现。
	CloseNoStatusReceived = websocket.CloseNoStatusReceived
	// 1006：同上。保留。且终端必须不在控制帧里设置作为状态码。它是指定给应用程序而非作为状态码使用的，用来指示连接非正常关闭，例如，没有发生或接收到关闭帧。
	CloseAbnormalClosure = websocket.CloseAbnormalClosure
	// 1007：表示终端因为接收到的数据没有消息类型而关闭连接。
	CloseInvalidFramePayloadData = websocket.CloseInvalidFramePayloadData
	// 1008：表示终端因为接收到的消息背离它的政策而关闭连接。这是一个通用的状态码，用在没有更合适的状态码或需要隐藏具体的政策细节时。
	ClosePolicyViolation = websocket.ClosePolicyViolation
	// 1009：表示终端因为接收到的消息太大以至于不能处理而关闭连接。
	CloseMessageTooBig = websocket.CloseMessageTooBig
	// 1010：表示客户端因为想和服务器协商一个或多个扩展，而服务器不在响应消息返回它（扩展）而关闭连接。需要的扩展列表应该出现在关闭帧的/reason/部分。注意，这个状态码不是由服务器使用，因为它会导致WebSocket握手失败。
	CloseMandatoryExtension = websocket.CloseMandatoryExtension
	// 1011：表示服务器因为遇到非预期的情况导致它不能完成请求而关闭连接。
	CloseInternalServerErr = websocket.CloseInternalServerErr
	CloseServiceRestart    = websocket.CloseServiceRestart
	CloseTryAgainLater     = websocket.CloseTryAgainLater
	// 1015：保留，且终端必须不在控制帧里设置作为状态码。它是指定用于应用程序希望用状态码来指示连接因为TLS握手失败而关闭。
	CloseTLSHandshake = websocket.CloseTLSHandshake
	/////////////////////

	// 以下是对库的错误状态码
	// 3000因网络错误而关闭
	CloseErrConn = 3000
	// 3000网络未打开, 需要重新连接网络
	CloseErrState = 3001
	// 因长时间没收到ping请求而关闭网络
	CloseErrTimeout = 3002

	// 以下是业务错误码不需要重试(4000-5000)
	// 4000因鉴权问题而关闭，需要重新登录服务器，不需进行重试
	CloseErrAuth = 4000
)

var (
	// 通用错误
	ErrNone = NewErrCode(200, "请求成功")

	// 其他错误码以两位进行分组，两位进行细分, 例如：1001

	// 以下错误码会触发关闭连接
	ErrCloseSys     = NewErrCode(CloseInternalServerErr, "系统升级中")
	ErrCloseAuth    = NewErrCode(CloseErrAuth, "您未登录，请登录")
	ErrCloseReLogin = NewErrCode(CloseErrAuth, "您已在其他地方登录，请重登录")
	ErrCloseProto   = NewErrCode(CloseUnsupportedData, "业务协议不正确")
	ErrCloseKick    = NewErrCode(CloseNormalClosure, "您已被踢下线")

	ErrSys      = NewErrCode(5000, "系统升级中")
	ErrAppProto = NewErrCode(5001, "协议不正确")
	ErrLogin    = NewErrCode(5101, "用户上线通知")
	ErrLogout   = NewErrCode(5102, "用户下线通知")
	ErrNotLogin = NewErrCode(5103, "用户不在线")
)
