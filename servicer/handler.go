package servicer

import "github.com/hacash/channelpay/protocol"

// 消息处理
func (s *Servicer) msgHandler(customer *Customer, msgobj protocol.Message, msgdata []byte) {

	switch msgobj.Type() {

	// 退出
	case protocol.MsgTypeLogout:
		customer.wsConn.Close() // 直接关闭连接
		break

	// 预查询支付
	case protocol.MsgTypeRequestPrequeryPayment:
		s.MsgHandlerRequestPrequeryPayment(customer, msgobj.(*protocol.MsgRequestPrequeryPayment))
		break

	// 确定发起支付
	case protocol.MsgTypeInitiatePayment:

		break
	}

}
