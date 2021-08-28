package servicer

import (
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
)

func (s *Servicer) connectHandler(ws *websocket.Conn) {

	// 创建客户连接
	customer := NewCustomer(ws)

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
					IsNonExistent: fields.CreateBool(false),
				}
				if customer.latestReconciliationBalanceBill != nil {
					billmsg.IsNonExistent = fields.CreateBool(true)
					billmsg.LastBill = customer.latestReconciliationBalanceBill
				}
				protocol.SendMsg(customer.wsConn, billmsg)
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
