package protocol

import "fmt"

const (
	LatestProtocolVersion uint16 = 1 // Latest protocol version number
)

const (
	// currency
	MsgTypeError uint8 = 0 // An error occurred

	// Channel routing sends data updates
	MsgTypePayRouteRequestServiceNodes      uint8 = 255 // Request node list
	MsgTypePayRouteResponseServiceNodes     uint8 = 254 // Response node list
	MsgTypePayRouteRequestNodeRelationship  uint8 = 253 // Request node connection relationship
	MsgTypePayRouteResponseNodeRelationship uint8 = 252 // Response node connection
	MsgTypePayRouteRequestUpdates           uint8 = 251 // Request update
	MsgTypePayRouteResponseUpdates          uint8 = 250 // Response update
	MsgTypePayRouteEndClose                 uint8 = 249 // Complete shutdown

	// Server send
	MsgTypeDisplacementOffline     uint8 = 1 // Remote login is offline
	MsgTypeLoginCheckLastestBill   uint8 = 2 // The server sends the latest statement
	MsgTypeResponsePrequeryPayment uint8 = 3 // Pre query payment information

	// Client send
	MsgTypeLogin                  uint8 = 4 // Customer login message
	MsgTypeLogout                 uint8 = 5 // Active client offline
	MsgTypeRequestPrequeryPayment uint8 = 6 // Pre query payment information
	MsgTypeInitiatePayment        uint8 = 7 // Initiate payment
	MsgTypeRelayInitiatePayment   uint8 = 8 // Relay node payment message

	// Payment related
	MsgTypeBroadcastChannelStatementProveBody uint8 = 9  // Broadcast statement
	MsgTypeBroadcastChannelStatementSignature uint8 = 10 // Broadcast channel payment signature
	MsgTypeBroadcastChannelStatementError     uint8 = 11 // Broadcast channel payment error
	MsgTypeBroadcastChannelStatementSuccessed uint8 = 12 // Broadcast channel payment completed successfully

	// Reconciliation related
	MsgTypeClientInitiateReconciliation  uint8 = 13 // Client initiated reconciliation
	MsgTypeServicerRespondReconciliation uint8 = 14 // Server response reconciliation

	// Client heartbeat package
	MsgTypeHeartbeat uint8 = 15

	/*

		MsgTypeRequestChannelPayCollectionSign               uint8 = 101 // Request collection signature from client
		MsgTypeResponseChannelPayCollectionSign              uint8 = 102 // Get signature
		MsgTypeRequestChannelPayPaymentSign                  uint8 = 103 // Request payment signature from client
		MsgTypeResponseChannelPayPaymentSign                 uint8 = 104 // Get signature
		MsgTypeSendChannelPayCompletedSignedBillToDownstream uint8 = 105 // Send complete bill to payment downstream

		MsgTypeResponseRemoteChannelPayment           uint8 = 107 // The remote payment is finally responded by the target terminal
		MsgTypeRequestRemoteChannelPayCollectionSign  uint8 = 108 // Request collection signature from remote
		MsgTypeResponseRemoteChannelPayCollectionSign uint8 = 109 // Remote signature reply
	*/

)

/**
 * 消息接口
 */
type Message interface {
	Type() uint8 // type
	Size() uint32
	Parse(buf []byte, seek uint32) (uint32, error)
	Serialize() ([]byte, error)         // serialize
	SerializeWithType() ([]byte, error) // serialize
}

/**
 * 解析消息
 */
func ParseMessage(buf []byte, seek uint32) (Message, error) {

	ty := buf[seek]
	var msg Message = nil

	// type
	switch ty {
	case MsgTypeError:
		msg = &MsgError{}
	case MsgTypeDisplacementOffline:
		msg = &MsgDisplacementOffline{}
	case MsgTypeLoginCheckLastestBill:
		msg = &MsgLoginCheckLastestBill{}

	case MsgTypeHeartbeat:
		msg = &MsgHeartbeat{}

	case MsgTypeLogin:
		msg = &MsgLogin{}
	case MsgTypeLogout:
		msg = &MsgCustomerLogout{}
	case MsgTypeRequestPrequeryPayment:
		msg = &MsgRequestPrequeryPayment{}
	case MsgTypeResponsePrequeryPayment:
		msg = &MsgResponsePrequeryPayment{}
	case MsgTypeInitiatePayment:
		msg = &MsgRequestInitiatePayment{}
	case MsgTypeRelayInitiatePayment:
		msg = &MsgRequestRelayInitiatePayment{}

	case MsgTypeBroadcastChannelStatementProveBody:
		msg = &MsgBroadcastChannelStatementProveBody{}
	case MsgTypeBroadcastChannelStatementSignature:
		msg = &MsgBroadcastChannelStatementSignature{}
	case MsgTypeBroadcastChannelStatementError:
		msg = &MsgBroadcastChannelStatementError{}
	case MsgTypeBroadcastChannelStatementSuccessed:
		msg = &MsgBroadcastChannelStatementSuccessed{}

	case MsgTypeClientInitiateReconciliation:
		msg = &MsgClientInitiateReconciliation{}
	case MsgTypeServicerRespondReconciliation:
		msg = &MsgServicerRespondReconciliation{}

	case MsgTypePayRouteRequestServiceNodes:
		msg = &MsgPayRouteRequestServiceNodes{}
	case MsgTypePayRouteResponseServiceNodes:
		msg = &MsgPayRouteResponseServiceNodes{}
	case MsgTypePayRouteRequestNodeRelationship:
		msg = &MsgPayRouteRequestNodeRelationship{}
	case MsgTypePayRouteResponseNodeRelationship:
		msg = &MsgPayRouteResponseNodeRelationship{}
	case MsgTypePayRouteRequestUpdates:
		msg = &MsgPayRouteRequestUpdates{}
	case MsgTypePayRouteResponseUpdates:
		msg = &MsgPayRouteResponseUpdates{}
	case MsgTypePayRouteEndClose:
		msg = &MsgPayRouteEndClose{}

	default:
		return nil, fmt.Errorf("Unsupported message type <%d>", ty)
	}

	// analysis
	_, e := msg.Parse(buf, seek+1)
	if e != nil {
		return nil, e
	}
	return msg, nil
}
