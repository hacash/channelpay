package protocol

import "github.com/hacash/node/websocket"

// 接受消息
func ReceiveMsg(wsconn *websocket.Conn) (Message, []byte, error) {
	var msgdata []byte = nil
	e := websocket.Message.Receive(wsconn, &msgdata)
	if e != nil {
		return nil, nil, e
	}
	// 解析消息错误，断开连接
	msgobj, err := ParseMessage(msgdata, 0)
	if err != nil {
		return nil, nil, err
	}
	// 成功
	return msgobj, msgdata, nil
}
