package servicer

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/node/websocket"
)

/**
 * 处理中继支付消息
 */
func (s *Servicer) dealRelayInitiatePayment(ws *websocket.Conn, msg *protocol.MsgRequestRelayInitiatePayment) error {

	sname := msg.IdentificationName.Value()
	cid := msg.ChannelId
	ourIsLeft := true

	// Query settlement channel table
	okfind := false
	s.settlenoderChgLock.RLock()
	nlist := s.settlenoder[sname]
	if nlist != nil {
		for _, v := range nlist {
			if v.ChannelId.Equal(msg.ChannelId) {
				okfind = true
				ourIsLeft = v.OurAddressIsLeft
				break
			}
		}
	}
	s.settlenoderChgLock.RUnlock()
	if !okfind {
		return fmt.Errorf("not find node of %s:%s", sname, hex.EncodeToString(cid))
	}

	side := chanpay.NewChannelSideById(cid)
	side.WsConn = ws
	e := s.InitializeChannelSide(side, nil, ourIsLeft)
	if e != nil {
		return e // Error initializing channelside
	}

	// Initialize node
	node := chanpay.NewRelayPayNodeConnect(sname, cid, ourIsLeft, side)

	// Transfer payment
	s.MsgHandlerRequestInitiatePayment(nil, node, &msg.InitPayMsg)

	// Listen for messages and wait
	msgch := make(chan protocol.Message, 1)
	subobj := side.SubscribeMessage(msgch)

	// wait for
	for {
		select {
		case <-subobj.Err():
			return nil
		case <-msgch:
			continue // wait for
		}
	}

}
