package protocol

import (
	"fmt"
	"github.com/hacash/node/websocket"
	"time"
)

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

// 接受消息
// timeoutsec 超时秒
func ReceiveMsgOfTimeout(wsconn *websocket.Conn, timeoutsec int) (Message, []byte, error) {

	// 管道
	var msgdata []byte = nil
	var msgobj Message = nil
	resErrChan := make(chan error, 2)

	// 超时
	ttk := time.AfterFunc(time.Duration(timeoutsec)*time.Second, func() {
		resErrChan <- fmt.Errorf("Receive message timeout") // 超时退出
	})

	// 读取消息
	go func() {
		var mdata []byte = nil
		e := websocket.Message.Receive(wsconn, &msgdata)
		if e != nil {
			resErrChan <- e
			return
		}
		// 解析消息错误，断开连接
		mobj, err := ParseMessage(msgdata, 0)
		if err != nil {
			resErrChan <- e
			return
		}
		// 停止超时回调
		ttk.Stop()
		// 成功
		msgdata, msgobj = mdata, mobj
		resErrChan <- nil
		return
	}()

	// 等待
	err := <-resErrChan
	if err == nil {
		return nil, nil, err
	}

	// 答复成功
	return msgobj, msgdata, nil
}
