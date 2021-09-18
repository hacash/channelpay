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
		if customer.IsRegistered == false {
			ws.Close() // 超时未注册，关闭
		}
	})

	// 循环读取消息
	for {
		// 读取消息，解析消息错误，断开连接
		msgobj, msgdata, err := protocol.ReceiveMsg(ws)
		if err != nil {
			break
		}
		// 首条消息必须为注册消息
		if customer.IsRegistered == false {
			if msgobj.Type() == protocol.MsgTypeLogin {
				tarobj := msgobj.(*protocol.MsgLogin)
				customer.DoRegister(tarobj.ChannelId, tarobj.CustomerAddress) // 执行注册
				old, _ := s.AddCustomerToPool(customer)                       // 添加到管理池
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
				// 继续接受消息
				continue
			} else {
				// 消息类型错误，直接退出
				break
			}
		}

		// 处理其它类型消息
		s.msgHandler(customer, msgobj, msgdata)

	}

	// 从管理池里移除
	s.RemoveCustomerFromPool(customer)

	// 断开连接
	ws.Close()
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
	}

	isLaunchPay := false

	// 如果 3 秒钟之内还未收到发起支付消息，则关闭连接
	time.AfterFunc(time.Second*4, func() {
		if isLaunchPay == false {
			ws.Close() // 超时未注册，关闭
		}
	})

	// 循环读取消息
	for {
		// 读取消息，解析消息错误，断开连接
		msgobj, _, err := protocol.ReceiveMsg(ws)
		if err != nil {
			break // 断开
		}

		mty := msgobj.Type()
		if mty == protocol.MsgTypeRequestLaunchRemoteChannelPayment {
			// 发起支付消息
			initpaymsg, ok := msgobj.(*protocol.MsgRequestLaunchRemoteChannelPayment)
			if !ok {
				errorReturn(fmt.Errorf("MsgRequestInitiatePayment format error"))
				break // 消息解析错误
			}
			isLaunchPay = true // 成功发起支付
			// 处理远程支付
			e := s.dealRemoteRelayPay(ws, initpaymsg)
			if e != nil {
				errorReturn(fmt.Errorf("DealRemotePay error: %s", e.Error()))
				break
			}
		}

	}

	// 断开连接
	ws.Close()
}
