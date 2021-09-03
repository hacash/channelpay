package protocol

import (
	"github.com/hacash/node/websocket"
)

// 发送消息
func SendMsg(wsconn *websocket.Conn, msg Message) error {
	bt, e := msg.SerializeWithType()
	if e != nil {
		return e
	}
	_, e = wsconn.Write(bt)
	return e
}

// 发送消息并取得回复
// timeoutsec 超时秒
func SendMsgForResponseTimeout(wsconn *websocket.Conn, msg Message, timeoutsec int) (Message, []byte, error) {

	// 发送消息
	e := SendMsg(wsconn, msg)
	if e != nil {
		return nil, nil, e
	}

	// 读取回复
	return ReceiveMsgOfTimeout(wsconn, timeoutsec)
}
