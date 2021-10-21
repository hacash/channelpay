package protocol

import "fmt"

const (
	LatestProtocolVersion uint16 = 1 // 最新的协议版本号
)

const (
	// 通用
	MsgTypeError uint8 = 0 // 发生错误

	// 通道路由下发数据更新
	MsgTypePayRouteRequestServiceNodes      uint8 = 255 // 请求节点列表
	MsgTypePayRouteResponseServiceNodes     uint8 = 254 // 响应节点列表
	MsgTypePayRouteRequestNodeRelationship  uint8 = 253 // 请求节点连接关系
	MsgTypePayRouteResponseNodeRelationship uint8 = 252 // 响应节点连接关系
	MsgTypePayRouteRequestUpdates           uint8 = 251 // 请求更新
	MsgTypePayRouteResponseUpdates          uint8 = 250 // 响应更新
	MsgTypePayRouteEndClose                 uint8 = 249 // 完成关闭

	// 服务端发送
	MsgTypeDisplacementOffline     uint8 = 1 // 异地登录被顶下线
	MsgTypeLoginCheckLastestBill   uint8 = 2 // 服务端发送最新对账单
	MsgTypeResponsePrequeryPayment uint8 = 3 // 预查询支付信息

	// 客户端发送
	MsgTypeLogin                  uint8 = 4 // 顾客登录消息
	MsgTypeLogout                 uint8 = 5 // 客户端主动下线
	MsgTypeRequestPrequeryPayment uint8 = 6 // 预查询支付信息
	MsgTypeInitiatePayment        uint8 = 7 // 发起支付
	MsgTypeRelayInitiatePayment   uint8 = 8 // 中继节点支付消息

	// 支付相关
	MsgTypeBroadcastChannelStatementProveBody uint8 = 9  // 广播对账单
	MsgTypeBroadcastChannelStatementSignature uint8 = 10 // 广播通道支付签名
	MsgTypeBroadcastChannelStatementError     uint8 = 11 // 广播通道支付错误
	MsgTypeBroadcastChannelStatementSuccessed uint8 = 12 // 广播通道支付成功完成

	// 对账相关
	MsgTypeClientInitiateReconciliation  uint8 = 13 // 客户端发起对账
	MsgTypeServicerRespondReconciliation uint8 = 14 // 服务端响应对账

	/*

		MsgTypeRequestChannelPayCollectionSign               uint8 = 101 // 向客户端请求收款签名
		MsgTypeResponseChannelPayCollectionSign              uint8 = 102 // 获得签名
		MsgTypeRequestChannelPayPaymentSign                  uint8 = 103 // 向客户端请求支付签名
		MsgTypeResponseChannelPayPaymentSign                 uint8 = 104 // 获得签名
		MsgTypeSendChannelPayCompletedSignedBillToDownstream uint8 = 105 // 发送完整票据给支付下游

		MsgTypeResponseRemoteChannelPayment           uint8 = 107 // 远程支付由目标终端最终响应
		MsgTypeRequestRemoteChannelPayCollectionSign  uint8 = 108 // 向远程请求收款签名
		MsgTypeResponseRemoteChannelPayCollectionSign uint8 = 109 // 远程签名回复
	*/

)

/**
 * 消息接口
 */
type Message interface {
	Type() uint8 // 类型
	Size() uint32
	Parse(buf []byte, seek uint32) (uint32, error)
	Serialize() ([]byte, error)         // 序列化
	SerializeWithType() ([]byte, error) // 序列化
}

/**
 * 解析消息
 */
func ParseMessage(buf []byte, seek uint32) (Message, error) {

	ty := buf[seek]
	var msg Message = nil

	// 类型
	switch ty {
	case MsgTypeError:
		msg = &MsgError{}
	case MsgTypeDisplacementOffline:
		msg = &MsgDisplacementOffline{}
	case MsgTypeLoginCheckLastestBill:
		msg = &MsgLoginCheckLastestBill{}

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

	// 解析
	_, e := msg.Parse(buf, seek+1)
	if e != nil {
		return nil, e
	}
	return msg, nil
}
