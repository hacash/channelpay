package protocol

import "fmt"

const (
	LatestProtocolVersion uint16 = 1 // 最新的协议版本号
)

const (
	// 通用
	MsgTypeError uint8 = 0 // 发生错误

	// 服务端发送
	MsgTypeDisplacementOffline     uint8 = 1 // 异地登录被顶下线
	MsgTypeLoginCheckLastestBill   uint8 = 2 // 服务端发送最新对账单
	MsgTypeResponsePrequeryPayment uint8 = 3 // 预查询支付信息

	// 客户端发送
	MsgTypeLogin                  uint8 = 101 // 顾客登录消息
	MsgTypeLogout                 uint8 = 102 // 客户端主动下线
	MsgTypeRequestPrequeryPayment uint8 = 103 // 预查询支付信息
	MsgTypeInitiatePayment        uint8 = 104 // 发起支付

	// 通道路由下发数据更新
	MsgTypePayRouteRequestServiceNodes      uint8 = 201 // 请求节点列表
	MsgTypePayRouteResponseServiceNodes     uint8 = 202 // 响应节点列表
	MsgTypePayRouteRequestNodeRelationship  uint8 = 203 // 请求节点连接关系
	MsgTypePayRouteResponseNodeRelationship uint8 = 204 // 响应节点连接关系
	MsgTypePayRouteRequestUpdates           uint8 = 205 // 请求更新
	MsgTypePayRouteResponseUpdates          uint8 = 206 // 响应更新
	MsgTypePayRouteEndClose                 uint8 = 207 // 完成关闭

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
	case MsgTypeLogin:
		msg = &MsgLogin{}
	case MsgTypeDisplacementOffline:
		msg = &MsgDisplacementOffline{}
	case MsgTypeLoginCheckLastestBill:
		msg = &MsgLoginCheckLastestBill{}
	case MsgTypeLogout:
		msg = &MsgCustomerLogout{}
	case MsgTypeRequestPrequeryPayment:
		msg = &MsgRequestPrequeryPayment{}
	case MsgTypeInitiatePayment:
		msg = &MsgInitiatePayment{}
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
