package servicer

import (
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
)

type Customer struct {
	IsRegistered bool // 是否已完成注册

	LanguageSet fields.StringMax255 // 语言设置 en_US zh_CN

	// 客户端长连接
	wsConn *websocket.Conn

	// 数据
	channelId fields.Bytes16 // 通道链 ID

	servicerAddress fields.Address // 服务节点地址
	customerAddress fields.Address // 客户地址

	// 最新的对账票据
	latestReconciliationBalanceBill channel.ReconciliationBalanceBill
}

func NewCustomer(ws *websocket.Conn) *Customer {
	return &Customer{
		IsRegistered: false,
		wsConn:       ws,
		channelId:    nil,
	}
}

// 执行注册
func (c *Customer) DoRegister(channelId fields.Bytes16, address fields.Address) {
	c.IsRegistered = true
	c.channelId = channelId
	c.customerAddress = address
}

// 被顶替下线
func (c *Customer) DoDisplacementOffline(newcur *Customer) {
	// 拷贝数据
	newcur.latestReconciliationBalanceBill = c.latestReconciliationBalanceBill
	newcur.servicerAddress = c.servicerAddress
	// 发送被顶替消息，被顶替者自动下线
	protocol.SendMsg(c.wsConn, &protocol.MsgDisplacementOffline{})
	// 关闭连接
	c.wsConn.Close()
}
