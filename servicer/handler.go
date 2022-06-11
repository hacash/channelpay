package servicer

import (
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
)

// Message processing
func (s *Servicer) msgHandler(customer *chanpay.Customer, msgobj protocol.Message) {

	switch msgobj.Type() {

	// sign out
	case protocol.MsgTypeLogout:
		customer.ChannelSide.WsConn.Close() // Close connection directly
		break

	// Pre query payment
	case protocol.MsgTypeRequestPrequeryPayment:
		s.MsgHandlerRequestPrequeryPayment(customer, msgobj.(*protocol.MsgRequestPrequeryPayment))
		break

	// Confirm to initiate payment
	case protocol.MsgTypeInitiatePayment:
		s.MsgHandlerRequestInitiatePayment(customer, nil, msgobj.(*protocol.MsgRequestInitiatePayment))
		break

	// Initiate reconciliation
	case protocol.MsgTypeClientInitiateReconciliation:
		s.MsgHandlerClientInitiateReconciliation(customer, msgobj.(*protocol.MsgClientInitiateReconciliation))
		break

	// Heartbeat 
	case protocol.MsgTypeHeartbeat:
		customer.UpdateLastestHeartbeatTime() // Keep heartbeat alive
		break

	}

}
