package servicer

import (
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
)

// 消息处理
func (s *Servicer) msgHandler(customer *chanpay.Customer, msgobj protocol.Message) {

	switch msgobj.Type() {

	// 退出
	case protocol.MsgTypeLogout:
		customer.ChannelSide.WsConn.Close() // 直接关闭连接
		break

	// 预查询支付
	case protocol.MsgTypeRequestPrequeryPayment:
		s.MsgHandlerRequestPrequeryPayment(customer, msgobj.(*protocol.MsgRequestPrequeryPayment))
		break

	// 确定发起支付
	case protocol.MsgTypeInitiatePayment:
		s.MsgHandlerRequestInitiatePayment(customer, nil, msgobj.(*protocol.MsgRequestInitiatePayment))
		break

	// 发起对账
	case protocol.MsgTypeClientInitiateReconciliation:
		s.MsgHandlerClientInitiateReconciliation(customer, msgobj.(*protocol.MsgClientInitiateReconciliation))
		break

	// 心跳包，忽略
	case protocol.MsgTypeHeartbeat:
		break

	}

}
