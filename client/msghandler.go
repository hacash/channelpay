package client

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
)

/**
 * 消息处理
 */
func (c *ChannelPayClient) startMsgHandler() {
	// 订阅消息处理，
	chanobj := make(chan protocol.Message, 1)
	subObj := c.user.upstreamSide.ChannelSide.SubscribeMessage(chanobj)
	c.user.msgSubObj = subObj
	// 循环处理消息
	go func() {
		//defer fmt.Println("ChannelPayUser.MsgHandler end")
		for {
			select {
			case v := <-chanobj:
				c.dealMsg(v)
			case <-subObj.Err():
				return // 订阅关闭
			}
		}
	}()
}

// 消息处理
func (c *ChannelPayClient) dealMsg(msg protocol.Message) {
	ty := msg.Type()
	switch ty {
	// 被顶下线
	case protocol.MsgTypeDisplacementOffline:
		c.user.Logout() // 退出登录
		// 向界面发出退出登录消息
		lgv := c.payui.Eval("Logout()") // 退出
		if DevDebug {
			fmt.Println("Logout() => ", lgv.String())
		}

		//ddd := dialog.NewInformation("Attention", "You have login at another place and this connection has been exited.", c.loginWindow)
		//ddd.SetOnClosed(func() {
		//	c.loginWindow.Close() // 关闭
		//})
		//ddd.Show() // 必须调用show否则事件响应有问题
	}

}
