package chanpay

import (
	"github.com/hacash/core/fields"
)

// Relay payment connection channel
type RelayPaySettleNoder struct {

	// Service name
	IdentificationName string
	ChannelId          fields.ChannelId
	OurAddressIsLeft   bool

	ChannelSide *ChannelSideConn
}

func NewRelayPayNodeConnect(name string, cid fields.ChannelId, ourIsLeft bool, side *ChannelSideConn) *RelayPaySettleNoder {
	return &RelayPaySettleNoder{
		IdentificationName: name,
		ChannelId:          cid,
		OurAddressIsLeft:   ourIsLeft,
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

// Check whether the collection channel is occupied
func (c *RelayPaySettleNoder) IsInBusinessExclusive() bool {
	return c.ChannelSide.IsInBusinessExclusive()
}

// Where state exclusive
func (c *RelayPaySettleNoder) StartBusinessExclusive() bool {
	return c.ChannelSide.StartBusinessExclusive()
}

// Remove state exclusivity
func (c *RelayPaySettleNoder) ClearBusinessExclusive() {
	c.ChannelSide.ClearBusinessExclusive()
}
