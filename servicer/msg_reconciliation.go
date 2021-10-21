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

	// 返回错误消息
	errorReturnString := func(err string) {
		errobj := &protocol.MsgError{
			ErrCode: 1,
		}
		errobj.ErrTip = fields.CreateStringMax65535(err)
		protocol.SendMsg(newcur.ChannelSide.WsConn, errobj)
	}

	// 最新账单
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

	// 检查并签署对账单
	sign, e := s.signmachine.CheckReconciliationFillNeedSignature(newbill, &msg.SelfSign)
	if e != nil || sign == nil {
		errorReturnString("CheckReconciliationFillNeedSignature error.")
		return // 签名检查失败
	}

	// 保存新的对账单
	s.billstore.UpdateStoreBalanceBill(billobj.GetChannelId(), newbill)
	newcur.ChannelSide.SetReconciliationBill(newbill)

	// 消息返回
	protocol.SendMsg(newcur.ChannelSide.WsConn, &protocol.MsgServicerRespondReconciliation{
		SelfSign: *sign,
	})
	// 成功
	return

}
