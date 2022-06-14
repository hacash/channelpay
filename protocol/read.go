package protocol

import (
	"fmt"
	"github.com/hacash/node/websocket"
	"time"
)

// Accept message
func ReceiveMsg(wsconn *websocket.Conn) (Message, []byte, error) {
	var msgdata []byte = nil
	e := websocket.Message.Receive(wsconn, &msgdata)
	if e != nil {
		return nil, nil, e
	}
	// Error parsing message, disconnect
	msgobj, err := ParseMessage(msgdata, 0)
	if err != nil {
		return nil, nil, err
	}
	// success
	return msgobj, msgdata, nil
}

// Accept message
// Timeoutsec timeout seconds
func ReceiveMsgOfTimeout(wsconn *websocket.Conn, timeoutsec int) (Message, []byte, error) {

	// The Conduit
	var msgdata []byte = nil
	var msgobj Message = nil
	resErrChan := make(chan error, 2)

	// overtime
	ttk := time.AfterFunc(time.Duration(timeoutsec)*time.Second, func() {
		resErrChan <- fmt.Errorf("Receive message timeout") // Timeout exit
	})

	// Read message
	go func() {
		var mdata []byte = nil
		e := websocket.Message.Receive(wsconn, &msgdata)
		if e != nil {
			resErrChan <- e
			return
		}
		// Error parsing message, disconnect
		mobj, err := ParseMessage(msgdata, 0)
		if err != nil {
			resErrChan <- e
			return
		}
		// Stop timeout callback
		ttk.Stop()
		// success
		msgdata, msgobj = mdata, mobj
		resErrChan <- nil
		return
	}()

	// wait for
	err := <-resErrChan
	if err != nil {
		return nil, nil, err
	}

	// Reply succeeded
	return msgobj, msgdata, nil
}
