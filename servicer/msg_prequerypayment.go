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

	// Return error message
	errorReturn := func(e error) {
		errobj := protocol.NewMsgResponsePrequeryPayment(1)
		errobj.ErrTip = fields.CreateStringMax65535(e.Error())
		protocol.SendMsg(newcur.ChannelSide.WsConn, errobj)
	}

	// Query payment path
	chanAddr := protocol.ChannelAccountAddress{}
	e := chanAddr.Parse(msg.PayeeChannelAddr.Value())
	if e != nil {
		// Address format error, sending error message
		errorReturn(e)
		return
	}

	// 不能自己转给自己
	if chanAddr.Address.Equal(newcur.ChannelSide.RemoteAddress) {
		errorReturn(fmt.Errorf("You can't transfer money to yourself."))
		return
	}

	// Whether the target is paid by the local service provider
	localServicerName := s.config.SelfIdentificationName
	localnode := s.payRouteMng.FindNodeByName(localServicerName)
	if localnode == nil {
		errorReturn(fmt.Errorf("Service Node <%s> not find in the routes list.", localServicerName))
		return
	}

	if chanAddr.CompareServiceName(localServicerName) {
		// Local Payment
		paysat := msg.PaySatoshi.GetRealSatoshi()
		forms := CreatePayPathFormsBySingleNodePath(localnode, &msg.PayAmount, &paysat)
		resmsg := protocol.NewMsgResponsePrequeryPayment(0)
		resmsg.Notes = fields.CreateStringMax65535("")
		resmsg.PathForms = forms
		// Message return
		protocol.SendMsg(newcur.ChannelSide.WsConn, resmsg)
		// success
		return
	}

	// Remote payment, query routing
	// Whether the target service provider exists
	tarNodeName := chanAddr.ServicerName.Value()
	targetnode := s.payRouteMng.FindNodeByName(tarNodeName)
	if targetnode == nil {
		errorReturn(fmt.Errorf("Target service Node <%s> not find in the routes list.", tarNodeName))
		return
	}

	// Query route
	pathResults, e := s.payRouteMng.SearchNodePath(localServicerName, tarNodeName)
	if e != nil {
		errorReturn(e)
		return
	}
	if len(pathResults) == 0 {
		// the path was not found
		errorReturn(fmt.Errorf("Can not find the pay routes path from node %s to %s.",
			localServicerName, tarNodeName))
		return
	}
	patsat := msg.PaySatoshi.GetRealSatoshi()
	forms := CreatePayPathForms(pathResults, &msg.PayAmount, &patsat) // 路径列表
	// news
	resmsg := protocol.NewMsgResponsePrequeryPayment(0)
	resmsg.Notes = fields.CreateStringMax65535("")
	resmsg.PathForms = forms
	// Message return
	protocol.SendMsg(newcur.ChannelSide.WsConn, resmsg)
	// success
	return

}
