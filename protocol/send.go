package protocol

import (
	"github.com/hacash/node/websocket"
)

// send message
func SendMsg(wsconn *websocket.Conn, msg Message) error {
	bt, e := msg.SerializeWithType()
	if e != nil {
		return e
	}
	_, e = wsconn.Write(bt)
	return e
}

// Send a message and get a reply
// Timeoutsec timeout seconds
func SendMsgForResponseTimeout(wsconn *websocket.Conn, msg Message, timeoutsec int) (Message, []byte, error) {

	// send message
	e := SendMsg(wsconn, msg)
	if e != nil {
		return nil, nil, e
	}

	// Read reply
	return ReceiveMsgOfTimeout(wsconn, timeoutsec)
}
