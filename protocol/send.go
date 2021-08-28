package protocol

import "github.com/hacash/node/websocket"

// 发送消息
func SendMsg(wsconn *websocket.Conn, msg Message) error {
	bt, e := msg.SerializeWithType()
	if e != nil {
		return e
	}
	_, e = wsconn.Write(bt)
	return e
}
