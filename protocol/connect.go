package protocol

import (
	"fmt"
	"github.com/hacash/node/websocket"
	"time"
)

// Initiate connection
// Send a message and get a reply
// Timeoutsec timeout seconds
func OpenConnectAndSendMsgForResponseTimeout(wsurl string, msg Message, timeoutsec int) (*websocket.Conn, Message, []byte, error) {

	var wsconn *websocket.Conn
	var errchan = make(chan error, 2)

	// overtime
	ttt := time.AfterFunc(time.Second*time.Duration(timeoutsec), func() {
		errchan <- fmt.Errorf("Dial %d second time out.")
	})

	go func() {
		// Initiate connection
		conn, e := websocket.Dial(wsurl, "", "http://127.0.0.1/")
		wsconn = conn // assignment
		ttt.Stop()
		errchan <- e
	}()

	// Waiting for response
	e := <-errchan
	if e != nil {
		return nil, nil, nil, e
	}

	// send message
	e = SendMsg(wsconn, msg)
	if e != nil {
		return nil, nil, nil, e
	}

	// Read reply
	msgobj, msgdata, e := ReceiveMsgOfTimeout(wsconn, timeoutsec)
	if e != nil {
		return nil, nil, nil, e
	}

	// complete
	return wsconn, msgobj, msgdata, nil
}

// connect
func OpenConnectAndSendMsg(wsurl string, msg Message) (*websocket.Conn, error) {

	// Initiate connection
	conn, e := websocket.Dial(wsurl, "", "http://127.0.0.1/")
	if e != nil {
		return nil, e
	}

	// send message
	e = SendMsg(conn, msg)
	if e != nil {
		return nil, e
	}

	// complete
	return conn, nil
}

// connect
func OpenConnect(wsurl string) (*websocket.Conn, error) {
	// Initiate connection
	return websocket.Dial(wsurl, "", "http://127.0.0.1/")
}
