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

	// 查询结算通道表
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
		return e // 初始化 ChannelSide 错误
	}

	// 初始化 node
	node := chanpay.NewRelayPayNodeConnect(sname, cid, ourIsLeft, side)

	// 调起支付
	s.MsgHandlerRequestInitiatePayment(nil, node, &msg.InitPayMsg)

	// 监听消息并等待
	msgch := make(chan protocol.Message, 1)
	subobj := side.SubscribeMessage(msgch)

	// 等待
	for {
		select {
		case <-subobj.Err():
			return nil
		case <-msgch:
			continue // 等待
		}
	}

}
