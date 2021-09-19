package chanpay

import (
	"github.com/hacash/core/fields"
)

// 中继支付连接通道
type RelayPaySettleNoder struct {

	// 服务名称
	identificationName string
	channelId          fields.ChannelID

	ChannelSide *ChannelSideConn
}

func NewRelayPayNodeConnect(name string, side *ChannelSideConn) *RelayPaySettleNoder {
	return &RelayPaySettleNoder{
		identificationName: name,
		ChannelSide:        side,
	}
}

func CreateChannelSideConnWrapForRelayPayNodeConnect(list []*RelayPaySettleNoder) ChannelSideConnListByCollectCapacity {
	var res = make([]ChannelSideConnWrap, len(list))
	for i, v := range list {
		res[i] = v
	}
	return res
}

func (c *RelayPaySettleNoder) GetChannelCapacityAmountForRemoteCollect() fields.Amount {
	return c.ChannelSide.GetChannelCapacityAmountOfOur()
}

// 检查收款通道是否被占用
func (c *RelayPaySettleNoder) IsInBusinessExclusive() bool {
	return c.ChannelSide.IsInBusinessExclusive()
}

// 其中状态独占
func (c *RelayPaySettleNoder) StartBusinessExclusive() bool {
	return c.ChannelSide.StartBusinessExclusive()
}

// 解除状态独占
func (c *RelayPaySettleNoder) ClearBusinessExclusive() {
	c.ChannelSide.ClearBusinessExclusive()
}
