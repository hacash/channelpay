package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
	"time"
)

/**
 * 发起支付
 */
func (s *Servicer) MsgHandlerRequestInitiatePayment(payuser *chanpay.Customer, upstreamSide *chanpay.RelayPaySettleNoder, msg *protocol.MsgRequestInitiatePayment) {

	// 创建side
	var upChannelSide *chanpay.ChannelSideConn = nil
	var downChannelSide *chanpay.ChannelSideConn = nil
	var originWsConn *websocket.Conn = nil
	var downInitPayMsg protocol.Message = nil
	if payuser != nil {
		upChannelSide = payuser.ChannelSide
		originWsConn = payuser.ChannelSide.WsConn
	} else if upstreamSide != nil {
		upChannelSide = upstreamSide.ChannelSide
		originWsConn = upstreamSide.ChannelSide.WsConn
	} else {
		return // 错误
	}

	// 返回错误消息
	var payins *chanpay.ChannelPayActionInstance = nil
	errorReturn := func(e error) {
		errmsg := &protocol.MsgBroadcastChannelStatementError{
			ErrCode: 1,
			ErrTip:  fields.CreateStringMax65535(e.Error()),
		}
		protocol.SendMsg(originWsConn, errmsg)
		if payins != nil {
			payins.Destroy() // 支付操作包销毁
		}
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
		localnode.Size()
		// 路由中继最多8层
		errorReturn(fmt.Errorf("Routing distance cannot be more than 8"))
		return
	}

	// 创建支付操作包
	payins = chanpay.NewChannelPayActionInstance()
	payins.SetLocalServicerNode(localnode)

	// 锁定和设置上游
	if payuser != nil {

		// 支付方通道独占
		if false == payuser.StartBusinessExclusive() {
			errorReturn(fmt.Errorf("User payment channel is being occupied, please try again later"))
			return
		}
		// 设置
		upChannelSide = payuser.ChannelSide
		payins.SetPayCustomer(payuser) // 设置上游通道

	} else if upstreamSide != nil {

		// 上游中继节点
		if false == upstreamSide.StartBusinessExclusive() {
			errorReturn(fmt.Errorf("Upstream side payment channel is being occupied, please try again later"))
			return
		}

		// 监听消息
		upstreamSide.ChannelSide.StartMessageListen()

		// 设置上游
		upChannelSide = upstreamSide.ChannelSide
		payins.SetUpstreamSide(upstreamSide) // 设置上游通道

	} else {
		errorReturn(fmt.Errorf("payuser and upstreamSide cannot both is nil."))
		return
	}

	// 检查上游支付通道容量
	paycap := upChannelSide.GetChannelCapacityAmountOfRemote()
	if paycap.LessThan(&msg.PayAmount) {
		// 上游通道资金过小不足以支付，检查暂时不包含手续费
		errorReturn(fmt.Errorf("Upstream channel capacity amount not enough."))
		return
	}

	// 查询下游
	targetServerIsOur := targetAddr.CompareServiceName(s.config.SelfIdentificationName)
	if targetServerIsOur {
		// 我是终端服务商，查询收款方是否在线
		// 取出收款目标地址的连接
		targetCuntomers := make([]*chanpay.Customer, 0)
		s.customerChgLock.RLock()
		for _, v := range s.customers {
			if v.ChannelSide.RemoteAddress.Equal(targetAddr.Address) {
				targetCuntomers = append(targetCuntomers, v)
			}
		}
		s.customerChgLock.RUnlock()

		// 是否有在线的客户端
		cusnum := len(targetCuntomers)
		if cusnum == 0 {
			errorReturn(fmt.Errorf("Target address %s is offline.", targetAddr.Address.ToReadable()))
			return
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
			errorReturn(fmt.Errorf("The payee channel is occupied. Please try again later"))
			return
		}

		// 检查收款方通道容量
		if chanwideamt.LessThan(&msg.PayAmount) {
			// 通道收款容量不足
			errorReturn(fmt.Errorf("Target address channel collect capacity %s insufficient.", chanwideamt.ToFinString()))
			return
		}

		// 向收款方发送的调起支付消息
		downInitPayMsg = msg

		// 设置收款方下游
		downChannelSide = receiveCustomer.ChannelSide
		payins.SetCollectCustomer(receiveCustomer)

	} else {
		// 我是中继节点，查询下一级
		var nextSerId = int64(0)
		for _, v := range msg.TargetPath.NodeIdPath {
			if nextSerId == -1 {
				nextSerId = int64(v) // 下一个节点
				break
			}
			if v == localnode.ID {
				nextSerId = -1 // 下一个就是
				continue
			}
		}
		if nextSerId == 0 {
			errorReturn(fmt.Errorf("Cannot find next servicer ID on target path."))
			return
		}
		// 查询
		nextNode := s.payRouteMng.FindNodeById(uint32(nextSerId))
		if nextNode == nil {
			errorReturn(fmt.Errorf("Cannot find next servicer of id %d.", nextSerId))
			return
		}
		// 检查通道
		var targetNn = nextNode.IdentificationName.Value()
		var targetRelayNodes []*chanpay.RelayPaySettleNoder = nil
		s.customerChgLock.RLock()
		targetRelayNodes = s.settlenoder[targetNn]
		s.customerChgLock.RUnlock()

		// 是否存在
		if targetRelayNodes == nil {
			errorReturn(fmt.Errorf("Target relay node %s is not find on configs.", targetNn))
			return
		}

		// 筛选最适合收款
		var tarokNode = targetRelayNodes[0]
		var chanwideamt = tarokNode.GetChannelCapacityAmountForRemoteCollect()
		if len(targetRelayNodes) > 1 {
			// 找出收款最大的通道容量
			for i := 1; i < len(targetRelayNodes); i++ {
				v := targetRelayNodes[i]
				if v.IsInBusinessExclusive() {
					continue // 通道占用
				}
				wideamt := v.GetChannelCapacityAmountForRemoteCollect()
				if wideamt.MoreThan(&chanwideamt) {
					chanwideamt = wideamt
					tarokNode = v
				}
			}
		}

		// 收款方通道独占
		if false == tarokNode.StartBusinessExclusive() {
			errorReturn(fmt.Errorf("The relay node channel is occupied. Please try again later"))
			return
		}

		// 检查收款方通道容量
		if chanwideamt.LessThan(&msg.PayAmount) {
			// 通道收款容量不足
			errorReturn(fmt.Errorf("Target address channel collect capacity %s insufficient.", chanwideamt.ToFinString()))
			return
		}

		// 向中继节点发起 ws 连接
		wsptl := "wss://"
		if s.config.DebugTest {
			wsptl = "ws://" // 开发测试
		}
		wsUrl := wsptl + nextNode.Gateway1.Value() + "/relaypay/connect"
		// 中继支付消息
		downInitPayMsg = &protocol.MsgRequestRelayInitiatePayment{
			InitPayMsg:         *msg,
			IdentificationName: fields.CreateStringMax255(localnode.IdentificationName.Value()),
			ChannelId:          tarokNode.ChannelSide.ChannelId,
		}
		// 发起连接并发送消息
		newconn, e := protocol.OpenConnect(wsUrl)
		if e != nil {
			errorReturn(fmt.Errorf("Connect relay node  %s error : %s.", wsUrl, e.Error()))
			return
		}

		// 连接赋值
		tarokNode.ChannelSide.WsConn = newconn
		// 监听消息
		tarokNode.ChannelSide.StartMessageListen()

		// 设置下游
		downChannelSide = tarokNode.ChannelSide
		payins.SetDownstreamSide(tarokNode)
	}

	if downInitPayMsg == nil {
		errorReturn(fmt.Errorf("downInitPayMsg is nil."))
		return
	}

	// 初始化
	e = payins.InitCreateEmptyBillDocumentsByInitPayMsg(msg)
	if e != nil {
		errorReturn(fmt.Errorf("InitCreateEmptyBillDocumentsByInitPayMsg error: %s.", e.Error()))
		return
	}

	// 设置支付成功回调
	payins.SetSuccessedBackCall(s.callbackPaymentSuccessed)

	// 监听上下游消息
	payins.StartOneSideMessageSubscription(true, upChannelSide)
	payins.StartOneSideMessageSubscription(false, downChannelSide)

	// 设置我必须签名的地址
	mustaddrs := []fields.Address{upChannelSide.OurAddress}
	if upChannelSide.OurAddress.NotEqual(downChannelSide.OurAddress) {
		mustaddrs = append(mustaddrs, downChannelSide.OurAddress) // 两个不同地址
	}
	payins.SetMustSignAddresses(mustaddrs)

	// 设置签名机
	payins.SetSignatureMachine(s.signmachine)

	// 如果是测试环境则打印日志
	if s.config.DebugTest {
		// 订阅日志，启动日志订阅
		logschan := make(chan *chanpay.PayActionLog, 2)
		go func() {
			for {
				log := <-logschan
				if log == nil || log.IsEnd {
					return // 订阅结束
				}
				// 显示日志
				fmt.Println(log.Content)
			}
		}()
		logschan <- &chanpay.PayActionLog{
			IsSuccess: true,
			Content:   fmt.Sprintf("---- new collecting %s at %s ----", msg.PayAmount.ToFinString(), time.Now().Format("2006-01-02 15:04:05")),
		}
		payins.SubscribeLogs(logschan) // 日志订阅
	}

	// 向下游发送的调起支付消息
	e = protocol.SendMsg(downChannelSide.WsConn, downInitPayMsg)
	if e != nil {
		errorReturn(fmt.Errorf("Send msg to receive customer error: %s", e.Error()))
		return
	}

	// OK 支付操作初始化成功
	return
}

/*

// 开始本地支付
func (s *Servicer) launchLocalPay(localnode *payroutes.PayRelayNode, payuser *chanpay.Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {

	// 取出收款目标地址的连接
	targetCuntomers := make([]*chanpay.Customer, 0)
	s.customerChgLock.RLock()
	for _, v := range s.customers {
		if v.ChannelSide.RemoteAddress.Equal(targetAddr.Address) {
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

*/

/*


// 开始远程支付
func (s *Servicer) launchRemotePay(localnode *payroutes.PayRelayNode, newcur *chanpay.Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {

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

*/
