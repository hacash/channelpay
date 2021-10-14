package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
	"time"
)

// 客户端连接
func (s *Servicer) connectCustomerHandler(ws *websocket.Conn) {

	// 创建客户连接
	customer := chanpay.NewCustomer(ws)

	// 如果 5 秒钟之内还未注册，则关闭连接
	time.AfterFunc(time.Second*5, func() {
		if customer.RegisteredID == 0 {
			ws.Close() // 超时未注册，关闭
		}
	})

	// 循环读取消息
	for {
		// 读取消息，解析消息错误，断开连接
		msgobj, _, err := protocol.ReceiveMsg(ws)
		if err != nil {
			break
		}
		// 首条消息必须为注册消息
		if customer.RegisteredID == 0 {
			if msgobj.Type() == protocol.MsgTypeLogin {
				tarobj := msgobj.(*protocol.MsgLogin)
				customer.DoRegister(tarobj.ChannelId, tarobj.CustomerAddress) // 执行注册
				// 找出旧连接
				old := s.FindCustomersByChannel(tarobj.ChannelId)
				// 处理旧的客户端连接
				if old != nil {
					// 从管理池里移除
					s.RemoveCustomerFromPool(old)
					// 拷贝数据，顶替下线
					old.DoDisplacementOffline(customer)
				} else {
					// 全新登录
					e := s.LoginNewCustomer(customer)
					if e != nil {
						// 登录失败，发送错误消息
						emsg := &protocol.MsgError{
							ErrCode: 1,
							ErrTip:  fields.CreateStringMax65535(e.Error()),
						}
						protocol.SendMsg(ws, emsg)
						// 断开连接
						break
					}
				}
				// 添加新连接
				s.AddCustomerToPool(customer)
				// 发送对账单消息
				billmsg := &protocol.MsgLoginCheckLastestBill{
					ProtocolVersion: fields.VarUint2(protocol.LatestProtocolVersion),
					BillIsExistent:  fields.CreateBool(false),
				}
				cusbill := customer.ChannelSide.GetReconciliationBill()
				if cusbill != nil {
					billmsg.BillIsExistent = fields.CreateBool(true)
					billmsg.LastBill = cusbill
				}
				protocol.SendMsg(customer.ChannelSide.WsConn, billmsg)
				// 继续接受消息，订阅消息
				s.dealOtherMessage(customer)
				// 成功返回
				break
			} else {
				// 消息类型错误，直接退出
				break
			}
		}

	}

	// 从管理池里移除
	s.RemoveCustomerFromPool(customer)

	// 断开连接
	ws.Close()
}

// 中继支付服务连接
func (s *Servicer) dealOtherMessage(customer *chanpay.Customer) {

	msgch := make(chan protocol.Message, 2)
	subobj := customer.ChannelSide.SubscribeMessage(msgch) // 订阅消息

	// 处理其它类型消息
	for {
		select {
		case <-subobj.Err(): // 消息订阅错误
			// 从管理池里移除
			s.RemoveCustomerFromPool(customer)
			// 断开连接
			customer.ChannelSide.WsConn.Close()
			return
		case msg := <-msgch:
			// 处理其他消息
			go s.msgHandler(customer, msg)
		}

	}

}

// 中继支付服务连接
func (s *Servicer) connectRelayPayHandler(ws *websocket.Conn) {

	// 返回错误消息
	errorReturn := func(e error) {
		errmsg := &protocol.MsgError{
			ErrCode: 0,
			ErrTip:  fields.CreateStringMax65535(e.Error()),
		}
		protocol.SendMsg(ws, errmsg)
		ws.Close() // 断开
	}

	isLaunchPay := false

	// 如果 5 秒钟之内还未收到发起支付消息，则关闭连接
	time.AfterFunc(time.Second*5, func() {
		if isLaunchPay == false {
			ws.Close() // 超时未注册，关闭
		}
	})

	// 读取消息，解析消息错误，断开连接
	msgobj, _, err := protocol.ReceiveMsg(ws)
	if err != nil {
		ws.Close()
		return // 断开
	}

	mty := msgobj.Type()
	if mty == protocol.MsgTypeRelayInitiatePayment {
		// 发起支付消息
		initpaymsg, ok := msgobj.(*protocol.MsgRequestRelayInitiatePayment)
		if !ok {
			errorReturn(fmt.Errorf("MsgRequestRelayInitiatePayment format error"))
			return // 消息解析错误
		}

		isLaunchPay = true                              // 标记已经调起
		e := s.dealRelayInitiatePayment(ws, initpaymsg) // 处理支付
		if e != nil {
			errorReturn(fmt.Errorf("DealRelayInitiatePayment error: %s", e.Error()))
			return // 错误
		}

	} else {

		// 只接受远程支付消息
		errorReturn(fmt.Errorf("Msg type error"))
		return // 错误

	}

}
