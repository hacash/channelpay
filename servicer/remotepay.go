package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/node/websocket"
)

/**
 * 处理远程支付
 */

// 开始远程中继支付
func (s *Servicer) dealRemoteRelayPay(wsconn *websocket.Conn, msg *protocol.MsgRequestLaunchRemoteChannelPayment) error {

	var e error

	// 目标收款地址
	targetAddr := &protocol.ChannelAccountAddress{}
	e = targetAddr.Parse(msg.PayeeChannelAddr.Value())
	if e != nil {
		return e
	}

	// 判断如果我就是最后一个节点
	localnode, e := s.GetLocalServiceNode()
	if e != nil {
		return e
	}

	nids := msg.TargetPath.NodeIdPath
	nlen := len(nids)
	if nlen < 2 {
		return fmt.Errorf("NodeIdPath len cannot less than 2.")
	}
	if nids[nlen-1] == localnode.ID && targetAddr.CompareServiceName(localnode.IdentificationName.Value()) {
		// 我就是最终终端
		// 查询连接客户端
		curcus, e := s.FindAndStartBusinessExclusiveWithOneCustomersByAddress(targetAddr.Address, &msg.PayAmount)
		if e != nil {
			return e // 地址不在线等错误
		}
		// 通道已经被锁定
		curcus.IsInBusinessExclusive()

	} else {
		return fmt.Errorf("msg format error.")
	}

	// 全部支付动作完成
	return nil
}
