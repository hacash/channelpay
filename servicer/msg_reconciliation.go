package servicer

import (
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
)

/**
 * 发起对账处理
 */
func (s *Servicer) MsgHandlerClientInitiateReconciliation(newcur *chanpay.Customer, msg *protocol.MsgClientInitiateReconciliation) {

	// Return error message
	errorReturnString := func(err string) {
		errobj := &protocol.MsgError{
			ErrCode: 1,
		}
		errobj.ErrTip = fields.CreateStringMax65535(err)
		protocol.SendMsg(newcur.ChannelSide.WsConn, errobj)
	}

	// Latest bill
	oldbill := newcur.ChannelSide.GetReconciliationBill()
	if oldbill == nil {
		errorReturnString("reconciliation bill bot find.")
		return
	}
	if oldbill.TypeCode() != channel.BillTypeCodeSimplePay {
		//fmt.Println("bill type is not BillTypeCodeSimplePay:", oldbill.TypeCode())
		errorReturnString("bill type is not BillTypeCodeSimplePay.")
		return
	}
	billobj := oldbill.(*channel.OffChainCrossNodeSimplePaymentReconciliationBill)
	newbill := billobj.ConvertToRealtimeReconciliation()

	// Check and sign the statement
	sign, e := s.signmachine.CheckReconciliationFillNeedSignature(newbill, &msg.SelfSign)
	if e != nil || sign == nil {
		errorReturnString("CheckReconciliationFillNeedSignature error.")
		return // signature check failed
	}

	// Save new statement
	s.billstore.UpdateStoreBalanceBill(billobj.GetChannelId(), newbill)
	newcur.ChannelSide.SetReconciliationBill(newbill)

	// Message return
	protocol.SendMsg(newcur.ChannelSide.WsConn, &protocol.MsgServicerRespondReconciliation{
		SelfSign: *sign,
	})
	// success
	return

}
