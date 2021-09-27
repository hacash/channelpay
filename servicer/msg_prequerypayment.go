package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
)

/**
 * 预查询支付处理
 */
func (s *Servicer) MsgHandlerRequestPrequeryPayment(newcur *chanpay.Customer, msg *protocol.MsgRequestPrequeryPayment) {

	// 返回错误消息
	errorReturn := func(e error) {
		errobj := protocol.NewMsgResponsePrequeryPayment(1)
		errobj.ErrTip = fields.CreateStringMax65535(e.Error())
		protocol.SendMsg(newcur.ChannelSide.WsConn, errobj)
	}

	// 查询支付路径
	chanAddr := protocol.ChannelAccountAddress{}
	e := chanAddr.Parse(msg.PayeeChannelAddr.Value())
	if e != nil {
		// 地址格式错误，发送错误信息
		errorReturn(e)
		return
	}

	// 不能自己转给自己
	if chanAddr.Address.Equal(newcur.ChannelSide.RemoteAddress) {
		errorReturn(fmt.Errorf("You can't transfer money to yourself."))
		return
	}

	// 目标是否为本地服务商支付
	localServicerName := s.config.SelfIdentificationName
	localnode := s.payRouteMng.FindNodeByName(localServicerName)
	if localnode == nil {
		errorReturn(fmt.Errorf("Service Node <%s> not find in the routes list.", localServicerName))
		return
	}

	if chanAddr.CompareServiceName(localServicerName) {
		// 本地支付
		forms := CreatePayPathFormsBySingleNodePath(localnode, &msg.PayAmount)
		resmsg := protocol.NewMsgResponsePrequeryPayment(0)
		resmsg.Notes = fields.CreateStringMax65535("")
		resmsg.PathForms = forms
		// 消息返回
		protocol.SendMsg(newcur.ChannelSide.WsConn, resmsg)
		// 成功
		return
	}

	// 远程支付，查询路由
	// 目标服务商是否存在
	targetnode := s.payRouteMng.FindNodeByName(chanAddr.ServicerName.Value())
	if targetnode == nil {
		errorReturn(fmt.Errorf("Target service Node <%s> not find in the routes list.", localServicerName))
		return
	}

	// 查询路由
	pathResults, e := s.payRouteMng.SearchNodePath(localServicerName, chanAddr.ServicerName.Value())
	if e != nil {
		errorReturn(e)
		return
	}
	if len(pathResults) == 0 {
		// 未找到路径
		errorReturn(fmt.Errorf("Can not find the pay routes path from node %s to %s.",
			localServicerName, targetnode.IdentificationName))
		return
	}
	forms := CreatePayPathForms(pathResults, &msg.PayAmount) // 路径列表
	// 消息
	resmsg := protocol.NewMsgResponsePrequeryPayment(0)
	resmsg.Notes = fields.CreateStringMax65535("")
	resmsg.PathForms = forms
	// 消息返回
	protocol.SendMsg(newcur.ChannelSide.WsConn, resmsg)
	// 成功
	return

}
