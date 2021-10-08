package protocol

import (
	"fmt"
	"github.com/hacash/node/websocket"
	"time"
)

// 发起连接
// 发送消息并取得回复
// timeoutsec 超时秒
func OpenConnectAndSendMsgForResponseTimeout(wsurl string, msg Message, timeoutsec int) (*websocket.Conn, Message, []byte, error) {

	var wsconn *websocket.Conn
	var errchan = make(chan error, 2)

	// 超时
	ttt := time.AfterFunc(time.Second*time.Duration(timeoutsec), func() {
		errchan <- fmt.Errorf("Dial %d second time out.")
	})

	go func() {
		// 发起连接
		conn, e := websocket.Dial(wsurl, "", "")
		wsconn = conn // 赋值
		ttt.Stop()
		errchan <- e
	}()

	// 等待响应
	e := <-errchan
	if e != nil {
		return nil, nil, nil, e
	}

	// 发送消息
	e = SendMsg(wsconn, msg)
	if e != nil {
		return nil, nil, nil, e
	}

	// 读取回复
	msgobj, msgdata, e := ReceiveMsgOfTimeout(wsconn, timeoutsec)
	if e != nil {
		return nil, nil, nil, e
	}

	// 完成
	return wsconn, msgobj, msgdata, nil
}

// 连接
func OpenConnectAndSendMsg(wsurl string, msg Message) (*websocket.Conn, error) {

	// 发起连接
	conn, e := websocket.Dial(wsurl, "", "")
	if e != nil {
		return nil, e
	}

	// 发送消息
	e = SendMsg(conn, msg)
	if e != nil {
		return nil, e
	}

	// 完成
	return conn, nil
}
