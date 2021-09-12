package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/payroutes"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
	"math/rand"
)

/**
 * 发起支付
 */
func (s *Servicer) MsgHandlerRequestInitiatePayment(payuser *Customer, upstreamSide *RelayPaySettleNoder, msg *protocol.MsgRequestInitiatePayment) {

	// 创建side

	var originWsConn *websocket.Conn = nil
	if payuser != nil {
		originWsConn = payuser.ChannelSide.wsConn
	} else if upstreamSide != nil {
		originWsConn = upstreamSide.ChannelSide.wsConn
	} else {
		return // 错误
	}

	// 返回错误消息
	errorReturn := func(e error) {
		errmsg := &protocol.MsgError{
			ErrCode: 0,
			ErrTip:  fields.CreateStringMax65535(e.Error()),
		}
		protocol.SendMsg(originWsConn, errmsg)
	}

	var e error

	// 目标收款地址
	targetAddr := &protocol.ChannelAccountAddress{}
	e = targetAddr.Parse(msg.PayeeChannelAddr.Value())
	if e != nil {
		errorReturn(e)
		return
	}

	// 本地节点
	localnode, e := s.GetLocalServiceNode()
	if e != nil {
		errorReturn(e)
		return
	}

	// 检查路由
	nids := msg.TargetPath.NodeIdPath
	if len(nids) > 8 {
		// 路由中继最多8层
		errorReturn(fmt.Errorf("Routing distance cannot be more than 8"))
		return
	}

	// 不区分本地还是远程支付

	if payuser != nil {
		// 支付方通道独占
		if false == payuser.StartBusinessExclusive() {
			errorReturn(fmt.Errorf("Payment channel is being occupied, please try again later"))
			return
		}
		defer payuser.ClearBusinessExclusive() // 支付结束时，或者发生错误时，终止独占状态
	}

	//if targetAddr.CompareServiceName(s.config.SelfIdentificationName) {
	//	e = s.launchLocalPay(localnode, payuser, msg, targetAddr)
	//} else {
	//	e = s.launchRemotePay(localnode, payuser, msg, targetAddr)
	//}

	// 返回错误
	if e != nil {
		errorReturn(e)
		return
	}
}

// 开始本地支付
func (s *Servicer) launchLocalPay(localnode *payroutes.PayRelayNode, payuser *Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {

	// 取出收款目标地址的连接
	targetCuntomers := make([]*Customer, 0)
	s.customerChgLock.RLock()
	for _, v := range s.customers {
		if v.ChannelSide.remoteAddress.Equal(targetAddr.Address) {
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
	var chanwideamt = receiveCustomer.GetChannelCapacityAmountForRemoteCollect()
	if cusnum > 1 {
		// 找出收款最大的通道容量
		for i := 1; i < len(targetCuntomers); i++ {
			v := targetCuntomers[i]
			if v.IsInBusinessExclusive() {
				continue // 通道占用
			}
			wideamt := v.GetChannelCapacityAmountForRemoteCollect()
			if wideamt.MoreThan(&chanwideamt) {
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
	fee := localnode.PredictFeeForPay(&msg.PayAmount)
	realpayamtwithfee, e := msg.PayAmount.Add(fee)
	if e != nil {
		return fmt.Errorf("InitiatePayment of add fee fail: %s", e.Error())
	}
	capamt := payuser.GetChannelCapacityAmountForRemotePay()
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
	msg1, _, e := protocol.SendMsgForResponseTimeout(receiveCustomer.ChannelSide.wsConn, smsg1, timeoutsec)
	if e != nil {
		return e // 返回错误
	}
	if msg1.Type() != protocol.MsgTypeResponseChannelPayCollectionSign {
		return fmt.Errorf("Collection customer signature failed")
	}
	msgobj1 := msg1.(*protocol.MsgResponseChannelPayCollectionSign)
	if msgobj1.ErrorCode > 0 {
		return fmt.Errorf("Collection customer sign Error: ", msgobj1.ErrorMsg.Value()) // 返回错误
	}
	// 判断签名地址
	e = bills.ChainPayment.FillSignByPosition(msgobj1.Sign)
	if e != nil {
		return fmt.Errorf("Fill need sign Error: ", e.Error()) // 返回错误
	}

	// 2. 本地服务节点签名， 5秒超时返回错误
	// 服务节点地址
	localnodeaddrs := make([]fields.Address, 0)
	localnodeaddrs = append(localnodeaddrs, payuser.GetServicerAddress())
	if payuser.GetServicerAddress().NotEqual(receiveCustomer.GetServicerAddress()) {
		localnodeaddrs = append(localnodeaddrs, receiveCustomer.GetServicerAddress())
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
	msg3, _, e := protocol.SendMsgForResponseTimeout(payuser.ChannelSide.wsConn, smsg3, timeoutsec)
	if e != nil {
		return e // 返回错误
	}
	if msg3.Type() != protocol.MsgTypeResponseChannelPayCollectionSign {
		return fmt.Errorf("Payment customer signature failed")
	}
	msgobj3 := msg3.(*protocol.MsgResponseChannelPayPaymentSign)
	if msgobj3.ErrorCode > 0 {
		return fmt.Errorf("Payment customer sign Error: ", msgobj3.ErrorMsg.Value()) // 返回错误
	}
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
	e = protocol.SendMsg(receiveCustomer.ChannelSide.wsConn, msg4)
	if e != nil {
		return fmt.Errorf("SendChannelPayCompletedSignedBillToDownstream Error: ", e.Error()) // 返回错误
	}

	// 5. 保存支付票据，清除通道独占状态，支付完成
	e = receiveCustomer.ChannelSide.UncheckSignSaveChannelPayReconciliationBalanceBill(bills)
	if e != nil {
		return fmt.Errorf("Receive Customer SaveChannelPayReconciliationBalanceBill Error: ", e.Error()) // 返回错误
	}
	e = payuser.ChannelSide.UncheckSignSaveChannelPayReconciliationBalanceBill(bills)
	if e != nil {
		return fmt.Errorf("Payment Customer SaveChannelPayReconciliationBalanceBill Error: ", e.Error()) // 返回错误
	}

	// 全部支付动作完成
	return nil
}

// 开始远程支付
func (s *Servicer) launchRemotePay(localnode *payroutes.PayRelayNode, newcur *Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {

	// 展开路由节点并检查路径是否有效
	nids := msg.TargetPath.NodeIdPath
	nlen := len(nids)
	if nlen < 2 {
		return fmt.Errorf("Node id path cannot less than 2.")
	}

	// 第一个节点必须是自己
	if nids[0] != localnode.ID {
		return fmt.Errorf("First node id need %d but got %d.",
			localnode.ID, nids[0])
	}

	// 展开
	pathnodes := make([]*payroutes.PayRelayNode, nlen)
	pathnodes[0] = localnode
	for i := 1; i < nlen; i++ {
		id := uint32(nids[i])
		nd := s.payRouteMng.FindNodeById(id)
		if nd == nil {
			return fmt.Errorf("Not find node id %d", id)
		}
		pathnodes[i] = nd
	}

	// 向下游节点发起连接
	nextnode := pathnodes[1]
	urlpto := "wss"
	if s.config.DebugTest {
		urlpto = "ws" // 测试环境
	}
	wsurl := fmt.Sprintf("%s://%s/relaypay/connect", urlpto, nextnode.Gateway1)

	// 连接并发送发起支付消息
	var msg1 = &protocol.MsgRequestLaunchRemoteChannelPayment{}
	msg1.CopyFromInitiatePayment(msg)
	wsconn, msgobjres, _, e := protocol.OpenConnectAndSendMsgForResponseTimeout(wsurl, msg1, 15)
	if e != nil {
		return e
	}

	return nil
}
