package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"math/rand"
)

/**
 * 发起支付
 */

func (s *Servicer) MsgHandlerRequestInitiatePayment(payuser *Customer, msg *protocol.MsgRequestInitiatePayment) {

	// 返回错误消息
	errorReturn := func(e error) {
		errmsg := &protocol.MsgError{
			ErrCode: 0,
			ErrTip:  fields.CreateStringMax65535(e.Error()),
		}
		protocol.SendMsg(payuser.wsConn, errmsg)
	}

	// 支付方通道独占
	if false == payuser.StartBusinessExclusive() {
		errorReturn(fmt.Errorf("Payment channel is being occupied, please try again later"))
		return
	}
	defer payuser.ClearBusinessExclusive() // 支付结束时，或者发生错误时，终止独占状态

	var e error

	// 目标收款地址
	targetAddr := &protocol.ChannelAccountAddress{}
	e = targetAddr.Parse(msg.PayeeChannelAddr.Value())
	if e != nil {
		errorReturn(e)
		return
	}

	// 本地还是远程支付
	if targetAddr.CompareServiceName(s.config.SelfIdentificationName) {
		e = s.localPay(payuser, msg, targetAddr)
	} else {
		e = s.remotePay(payuser, msg, targetAddr)
	}

	// 返回错误
	if e != nil {
		errorReturn(e)
		return
	}
}

// 开始本地支付
func (s *Servicer) localPay(payuser *Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {

	// 取出收款目标地址的连接
	targetCuntomers := make([]*Customer, 0)
	s.customerChgLock.RLock()
	for _, v := range s.customers {
		if v.customerAddress.Equal(targetAddr.Address) {
			targetCuntomers = append(targetCuntomers, v)
		}
	}
	s.customerChgLock.RUnlock()

	// 是否有在线的客户端
	cusnum := len(targetCuntomers)
	if cusnum == 0 {
		return fmt.Errorf("Target address %s is offline.", targetAddr.Address.ToReadable())
	}

	// 筛选最适合收款
	var receiveCustomer = targetCuntomers[0]
	var chanwideamt = receiveCustomer.GetChannelCapacityAmountForCollect()
	if cusnum > 1 {
		// 找出收款最大的通道容量
		for i := 1; i < len(targetCuntomers); i++ {
			v := targetCuntomers[i]
			if v.IsInBusinessExclusive() {
				continue // 通道占用
			}
			wideamt := v.GetChannelCapacityAmountForCollect()
			if wideamt.MoreThan(chanwideamt) {
				chanwideamt = wideamt
				receiveCustomer = v
			}
		}
	}

	// 收款方通道独占
	if false == receiveCustomer.StartBusinessExclusive() {
		return fmt.Errorf("The payee channel is occupied. Please try again later")
	}
	defer receiveCustomer.ClearBusinessExclusive() // 清除独占

	// 检查收款方通道容量
	if chanwideamt.LessThan(&msg.PayAmount) {
		// 通道收款容量不足
		return fmt.Errorf("Target address channel collect capacity %s insufficient.", chanwideamt.ToFinString())
	}

	// 检查支付方通道容量
	localnode, e := s.GetLocalServiceNode()
	if e != nil {
		return e
	}
	fee := localnode.PredictFeeForPay(&msg.PayAmount)
	realpayamtwithfee, e := msg.PayAmount.Add(fee)
	if e != nil {
		return fmt.Errorf("InitiatePayment of add fee fail: %s", e.Error())
	}
	capamt := payuser.GetChannelCapacityAmountForPay()
	if capamt.LessThan(realpayamtwithfee) {
		// 支付通道余额不足
		return fmt.Errorf("Insufficient payment channel balance, need %s but got %s",
			realpayamtwithfee.ToFinString(), capamt.ToFinString())
	}

	// 所有状态检查完毕，创建票据，开启支付
	bills, e := s.CreateChannelPayTransferTransactionForLocalPay(
		payuser, receiveCustomer, &msg.PayAmount, realpayamtwithfee, msg.OrderNoteHashHalfChecker,
	)
	if e != nil {
		return fmt.Errorf("CreateChannelPayTransferTransactionForLocalPay Error: %s", e.Error())
	}

	// 将票据发送至收付双方签名
	timeoutsec := 5                                   // 5秒超时
	operationnumber := fields.VarUint8(rand.Uint64()) // 流水号

	// 1. 首先发给收款方签名， 5秒超时返回错误
	smsg1 := &protocol.MsgRequestChannelPayCollectionSign{
		OperationNum: operationnumber,
		Bills:        bills,
	}
	// 等待收款方签名
	msg1, _, e := protocol.SendMsgForResponseTimeout(receiveCustomer.wsConn, smsg1, timeoutsec)
	if e != nil {
		return e // 返回错误
	}
	if msg1.Type() != protocol.MsgTypeResponseChannelPayCollectionSign {
		return fmt.Errorf("Collection customer signature failed")
	}
	msgobj1 := msg1.(*protocol.MsgResponseChannelPayCollectionSign)
	// 判断签名地址
	e = bills.ChainPayment.FillSignByPosition(msgobj1.Sign)
	if e != nil {
		return fmt.Errorf("Fill need sign Error: ", e.Error()) // 返回错误
	}

	// 2. 本地服务节点签名， 5秒超时返回错误
	// 服务节点地址
	localnodeaddrs := make([]fields.Address, 0)
	localnodeaddrs = append(localnodeaddrs, payuser.servicerAddress)
	if payuser.servicerAddress.NotEqual(receiveCustomer.servicerAddress) {
		localnodeaddrs = append(localnodeaddrs, receiveCustomer.servicerAddress)
	}
	// 服务商签名
	nodesigns, e := s.signmachine.CheckPaydocumentAndFillNeedSignature(bills, localnodeaddrs)
	if e != nil {
		return fmt.Errorf("Fill need sign Error: ", e.Error()) // 返回错误
	}

	// 3. 付款方签名， 5秒超时返回错误
	smsg3 := &protocol.MsgRequestChannelPayPaymentSign{
		OperationNum: operationnumber,
		Bills:        bills,
	}
	msg3, _, e := protocol.SendMsgForResponseTimeout(payuser.wsConn, smsg3, timeoutsec)
	if e != nil {
		return e // 返回错误
	}
	if msg3.Type() != protocol.MsgTypeResponseChannelPayCollectionSign {
		return fmt.Errorf("Payment customer signature failed")
	}
	msgobj3 := msg3.(*protocol.MsgResponseChannelPayPaymentSign)
	// 判断签名地址
	e = bills.ChainPayment.FillSignByPosition(msgobj3.Sign)
	if e != nil {
		return fmt.Errorf("Fill need sign Error: ", e.Error()) // 返回错误
	}

	// 4. 验证全部签名，发送所有签名至收款方
	e = bills.ChainPayment.CheckMustAddressAndSigns()
	if e != nil {
		return fmt.Errorf("Check must signs Error: ", e.Error()) // 返回错误
	}
	// 发送全部签名至收款方
	signlist := fields.CreateEmptySignListMax255()
	signlist.Append(msgobj3.Sign)
	for _, v := range nodesigns.Signs {
		signlist.Append(v) // 本地节点签名
	}
	msg4 := &protocol.MsgSendChannelPayCompletedSignaturesToDownstream{
		OperationNum: operationnumber,
		AllSigns:     *signlist,
	}
	e = protocol.SendMsg(receiveCustomer.wsConn, msg4)
	if e != nil {
		return fmt.Errorf("SendChannelPayCompletedSignedBillToDownstream Error: ", e.Error()) // 返回错误
	}

	// 5. 保存支付票据，清除通道独占状态，支付完成
	e = receiveCustomer.UncheckSignSaveChannelPayReconciliationBalanceBill(bills)
	if e != nil {
		return fmt.Errorf("Receive Customer SaveChannelPayReconciliationBalanceBill Error: ", e.Error()) // 返回错误
	}
	e = payuser.UncheckSignSaveChannelPayReconciliationBalanceBill(bills)
	if e != nil {
		return fmt.Errorf("Payment Customer SaveChannelPayReconciliationBalanceBill Error: ", e.Error()) // 返回错误
	}

	// 全部支付动作完成
	return nil
}

// 开始远程支付
func (s *Servicer) remotePay(newcur *Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {
	return nil
}
