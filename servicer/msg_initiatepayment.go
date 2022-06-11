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

	// Create side
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
		return // error
	}

	// Return error message
	var payins *chanpay.ChannelPayActionInstance = nil
	errorReturn := func(e error) {
		errmsg := &protocol.MsgBroadcastChannelStatementError{
			ErrCode: 1,
			ErrTip:  fields.CreateStringMax65535(e.Error()),
		}
		protocol.SendMsg(originWsConn, errmsg)
		if payins != nil {
			payins.Destroy() // Payment operation package destruction
		}
	}

	var e error

	// Target collection address
	targetAddr := &protocol.ChannelAccountAddress{}
	e = targetAddr.Parse(msg.PayeeChannelAddr.Value())
	if e != nil {
		errorReturn(e)
		return
	}

	// Local node
	localnode, e := s.GetLocalServiceNode()
	if e != nil {
		errorReturn(e)
		return
	}

	// Check routing
	nids := msg.TargetPath.NodeIdPath
	if len(nids) > 8 {
		localnode.Size()
		// Up to 8 layers of routing relay
		errorReturn(fmt.Errorf("Routing distance cannot be more than 8"))
		return
	}

	// Create payment operation package
	payins = chanpay.NewChannelPayActionInstance()
	payins.SetLocalServicerNode(localnode)

	// Lock and set upstream
	if payuser != nil {

		// Exclusive channel of payer
		if false == payuser.StartBusinessExclusive() {
			errorReturn(fmt.Errorf("User payment channel is being occupied, please try again later"))
			return
		}
		// set up
		upChannelSide = payuser.ChannelSide
		payins.SetPayCustomer(payuser) // Set up upstream channel

	} else if upstreamSide != nil {

		// Upstream relay node
		if false == upstreamSide.StartBusinessExclusive() {
			errorReturn(fmt.Errorf("Upstream side payment channel is being occupied, please try again later"))
			return
		}

		// Listen for messages
		upstreamSide.ChannelSide.StartMessageListen()

		// Set upstream
		upChannelSide = upstreamSide.ChannelSide
		payins.SetUpstreamSide(upstreamSide) // Set up upstream channel

	} else {
		errorReturn(fmt.Errorf("payuser and upstreamSide cannot both is nil."))
		return
	}

	// Check the capacity of upstream payment channel
	paycap := upChannelSide.GetChannelCapacityAmountOfRemote()
	if paycap.LessThan(&msg.PayAmount) {
		// The funds of the upstream channel are too small to pay, and the inspection does not include the handling fee for the time being
		errorReturn(fmt.Errorf("Upstream channel capacity amount not enough."))
		return
	}

	// Query downstream
	targetServerIsOur := targetAddr.CompareServiceName(s.config.SelfIdentificationName)
	if targetServerIsOur {
		// 我是终端服务商，查询收款方是否在线
		// Get the connection of the collection destination address
		targetCuntomers := make([]*chanpay.Customer, 0)
		s.customerChgLock.RLock()
		for _, v := range s.customers {
			if v.ChannelSide.RemoteAddress.Equal(targetAddr.Address) {
				targetCuntomers = append(targetCuntomers, v)
			}
		}
		s.customerChgLock.RUnlock()

		// Is there an online client
		cusnum := len(targetCuntomers)
		if cusnum == 0 {
			errorReturn(fmt.Errorf("Target address %s is offline.", targetAddr.Address.ToReadable()))
			return
		}

		// Filter the most suitable collection
		var receiveCustomer = targetCuntomers[0]
		var chanwideamt = receiveCustomer.GetChannelCapacityAmountForRemoteCollect()
		if cusnum > 1 {
			// Find out the maximum channel capacity for collection
			for i := 1; i < len(targetCuntomers); i++ {
				v := targetCuntomers[i]
				if v.IsInBusinessExclusive() {
					continue // Channel occupancy
				}
				wideamt := v.GetChannelCapacityAmountForRemoteCollect()
				if wideamt.MoreThan(&chanwideamt) {
					chanwideamt = wideamt
					receiveCustomer = v
				}
			}
		}

		// Exclusive channel of payee
		if false == receiveCustomer.StartBusinessExclusive() {
			errorReturn(fmt.Errorf("The payee channel is occupied. Please try again later"))
			return
		}

		// Check the channel capacity of the payee
		if chanwideamt.LessThan(&msg.PayAmount) {
			// Insufficient channel collection capacity
			errorReturn(fmt.Errorf("Target address channel collect capacity %s insufficient.", chanwideamt.ToFinString()))
			return
		}

		// Transfer up payment message sent to payee
		downInitPayMsg = msg

		// Set payee downstream
		downChannelSide = receiveCustomer.ChannelSide
		payins.SetCollectCustomer(receiveCustomer)

	} else {
		// I am the relay node. Query the next level
		var nextSerId = int64(0)
		for _, v := range msg.TargetPath.NodeIdPath {
			if nextSerId == -1 {
				nextSerId = int64(v) // Next node
				break
			}
			if v == localnode.ID {
				nextSerId = -1 // Next is
				continue
			}
		}
		if nextSerId == 0 {
			errorReturn(fmt.Errorf("Cannot find next servicer ID on target path."))
			return
		}
		// query
		nextNode := s.payRouteMng.FindNodeById(uint32(nextSerId))
		if nextNode == nil {
			errorReturn(fmt.Errorf("Cannot find next servicer of id %d.", nextSerId))
			return
		}
		// access for inspection
		var targetNn = nextNode.IdentificationName.Value()
		var targetRelayNodes []*chanpay.RelayPaySettleNoder = nil
		s.customerChgLock.RLock()
		targetRelayNodes = s.settlenoder[targetNn]
		s.customerChgLock.RUnlock()

		// Whether it exists
		if targetRelayNodes == nil {
			errorReturn(fmt.Errorf("Target relay node %s is not find on configs.", targetNn))
			return
		}

		// Filter the most suitable collection
		var tarokNode = targetRelayNodes[0]
		var chanwideamt = tarokNode.GetChannelCapacityAmountForRemoteCollect()
		if len(targetRelayNodes) > 1 {
			// Find out the maximum channel capacity for collection
			for i := 1; i < len(targetRelayNodes); i++ {
				v := targetRelayNodes[i]
				if v.IsInBusinessExclusive() {
					continue // Channel occupancy
				}
				wideamt := v.GetChannelCapacityAmountForRemoteCollect()
				if wideamt.MoreThan(&chanwideamt) {
					chanwideamt = wideamt
					tarokNode = v
				}
			}
		}

		// Exclusive channel of payee
		if false == tarokNode.StartBusinessExclusive() {
			errorReturn(fmt.Errorf("The relay node channel is occupied. Please try again later"))
			return
		}

		// Check the channel capacity of the payee
		if chanwideamt.LessThan(&msg.PayAmount) {
			// Insufficient channel collection capacity
			errorReturn(fmt.Errorf("Target address channel collect capacity %s insufficient.", chanwideamt.ToFinString()))
			return
		}

		// Initiate WS connection to relay node
		wsptl := "wss://"
		if s.config.DebugTest {
			wsptl = "ws://" // 开发测试
		}
		wsUrl := wsptl + nextNode.Gateway1.Value() + "/relaypay/connect"
		// Relay payment message
		downInitPayMsg = &protocol.MsgRequestRelayInitiatePayment{
			InitPayMsg:         *msg,
			IdentificationName: fields.CreateStringMax255(localnode.IdentificationName.Value()),
			ChannelId:          tarokNode.ChannelSide.ChannelId,
		}
		// Initiate connection and send message
		newconn, e := protocol.OpenConnect(wsUrl)
		if e != nil {
			errorReturn(fmt.Errorf("Connect relay node  %s error : %s.", wsUrl, e.Error()))
			return
		}

		// Connection assignment
		tarokNode.ChannelSide.WsConn = newconn
		// Listen for messages
		tarokNode.ChannelSide.StartMessageListen()

		// Set downstream
		downChannelSide = tarokNode.ChannelSide
		payins.SetDownstreamSide(tarokNode)
	}

	if downInitPayMsg == nil {
		errorReturn(fmt.Errorf("downInitPayMsg is nil."))
		return
	}

	// initialization
	e = payins.InitCreateEmptyBillDocumentsByInitPayMsg(msg)
	if e != nil {
		errorReturn(fmt.Errorf("InitCreateEmptyBillDocumentsByInitPayMsg error: %s.", e.Error()))
		return
	}

	// Set payment success callback
	payins.SetSuccessedBackCall(s.callbackPaymentSuccessed)

	// Listen for upstream and downstream messages
	payins.StartOneSideMessageSubscription(true, upChannelSide)
	payins.StartOneSideMessageSubscription(false, downChannelSide)

	// Set the address I must sign
	mustaddrs := []fields.Address{upChannelSide.OurAddress}
	if upChannelSide.OurAddress.NotEqual(downChannelSide.OurAddress) {
		mustaddrs = append(mustaddrs, downChannelSide.OurAddress) // Two different addresses
	}
	payins.SetMustSignAddresses(mustaddrs)

	// Setting up a signer
	payins.SetSignatureMachine(s.signmachine)

	// Print log if test environment
	if s.config.DebugTest {
		// Log subscription, start log subscription
		logschan := make(chan *chanpay.PayActionLog, 2)
		go func() {
			for {
				log := <-logschan
				if log == nil || log.IsEnd {
					return // End of subscription
				}
				// Show log
				fmt.Println(log.Content)
			}
		}()
		logschan <- &chanpay.PayActionLog{
			IsSuccess: true,
			Content:   fmt.Sprintf("---- new collecting %s at %s ----", msg.PayAmount.ToFinString(), time.Now().Format("2006-01-02 15:04:05")),
		}
		payins.SubscribeLogs(logschan) // Log subscription
	}

	// Call up payment message sent to downstream
	e = protocol.SendMsg(downChannelSide.WsConn, downInitPayMsg)
	if e != nil {
		errorReturn(fmt.Errorf("Send msg to receive customer error: %s", e.Error()))
		return
	}

	// OK payment operation initialized successfully
	return
}

/*

// Start local payment
func (s *Servicer) launchLocalPay(localnode *payroutes.PayRelayNode, payuser *chanpay.Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {

	// Get the connection of the collection destination address
	targetCuntomers := make([]*chanpay.Customer, 0)
	s.customerChgLock.RLock()
	for _, v := range s.customers {
		if v.ChannelSide.RemoteAddress.Equal(targetAddr.Address) {
			targetCuntomers = append(targetCuntomers, v)
		}
	}
	s.customerChgLock.RUnlock()

	// Is there an online client
	cusnum := len(targetCuntomers)
	if cusnum == 0 {
		return fmt.Errorf("Target address %s is offline.", targetAddr.Address.ToReadable())
	}

	// Filter the most suitable collection
	var receiveCustomer = targetCuntomers[0]
	var chanwideamt = receiveCustomer.GetChannelCapacityAmountForRemoteCollect()
	if cusnum > 1 {
		// Find out the maximum channel capacity for collection
		for i := 1; i < len(targetCuntomers); i++ {
			v := targetCuntomers[i]
			if v.IsInBusinessExclusive() {
				continue // Channel occupancy
			}
			wideamt := v.GetChannelCapacityAmountForRemoteCollect()
			if wideamt.MoreThan(&chanwideamt) {
				chanwideamt = wideamt
				receiveCustomer = v
			}
		}
	}

	// Exclusive channel of payee
	if false == receiveCustomer.StartBusinessExclusive() {
		return fmt.Errorf("The payee channel is occupied. Please try again later")
	}
	defer receiveCustomer.ClearBusinessExclusive() // Clear exclusive

	// Check the channel capacity of the payee
	if chanwideamt.LessThan(&msg.PayAmount) {
		// Insufficient channel collection capacity
		return fmt.Errorf("Target address channel collect capacity %s insufficient.", chanwideamt.ToFinString())
	}

	// Check the channel capacity of the payer
	fee := localnode.PredictFeeForPay(&msg.PayAmount)
	realpayamtwithfee, e := msg.PayAmount.Add(fee)
	if e != nil {
		return fmt.Errorf("InitiatePayment of add fee fail: %s", e.Error())
	}
	capamt := payuser.GetChannelCapacityAmountForRemotePay()
	if capamt.LessThan(realpayamtwithfee) {
		// Insufficient payment channel balance
		return fmt.Errorf("Insufficient payment channel balance, need %s but got %s",
			realpayamtwithfee.ToFinString(), capamt.ToFinString())
	}

	// After all status checks are completed, create bills and start payment
	bills, e := s.CreateChannelPayTransferTransactionForLocalPay(
		payuser, receiveCustomer, &msg.PayAmount, realpayamtwithfee, msg.OrderNoteHashHalfChecker,
	)
	if e != nil {
		return fmt.Errorf("CreateChannelPayTransferTransactionForLocalPay Error: %s", e.Error())
	}

	// Send the bill to both parties for signature
	timeoutsec := 5                                   // 5秒超时
	operationnumber := fields.VarUint8(rand.Uint64()) // 流水号

	// 1. first send it to the payee for signature, and an error is returned after 5 seconds
	smsg1 := &protocol.MsgRequestChannelPayCollectionSign{
		OperationNum: operationnumber,
		Bills:        bills,
	}
	// Waiting for signature of payee
	msg1, _, e := protocol.SendMsgForResponseTimeout(receiveCustomer.ChannelSide.wsConn, smsg1, timeoutsec)
	if e != nil {
		return e // Return error
	}
	if msg1.Type() != protocol.MsgTypeResponseChannelPayCollectionSign {
		return fmt.Errorf("Collection customer signature failed")
	}
	msgobj1 := msg1.(*protocol.MsgResponseChannelPayCollectionSign)
	if msgobj1.ErrorCode > 0 {
		return fmt.Errorf("Collection customer sign Error: ", msgobj1.ErrorMsg.Value()) // 返回错误
	}
	// Judge signature address
	e = bills.ChainPayment.FillSignByPosition(msgobj1.Sign)
	if e != nil {
		return fmt.Errorf("Fill need sign Error: ", e.Error()) // 返回错误
	}

	// 2. the local service node signs, and returns an error after 5 seconds
	// Service node address
	localnodeaddrs := make([]fields.Address, 0)
	localnodeaddrs = append(localnodeaddrs, payuser.GetServicerAddress())
	if payuser.GetServicerAddress().NotEqual(receiveCustomer.GetServicerAddress()) {
		localnodeaddrs = append(localnodeaddrs, receiveCustomer.GetServicerAddress())
	}
	// Service provider signature
	nodesigns, e := s.signmachine.CheckPaydocumentAndFillNeedSignature(bills, localnodeaddrs)
	if e != nil {
		return fmt.Errorf("Fill need sign Error: ", e.Error()) // 返回错误
	}

	// 3. the payer signs and returns an error after 5 seconds
	smsg3 := &protocol.MsgRequestChannelPayPaymentSign{
		OperationNum: operationnumber,
		Bills:        bills,
	}
	msg3, _, e := protocol.SendMsgForResponseTimeout(payuser.ChannelSide.wsConn, smsg3, timeoutsec)
	if e != nil {
		return e // Return error
	}
	if msg3.Type() != protocol.MsgTypeResponseChannelPayCollectionSign {
		return fmt.Errorf("Payment customer signature failed")
	}
	msgobj3 := msg3.(*protocol.MsgResponseChannelPayPaymentSign)
	if msgobj3.ErrorCode > 0 {
		return fmt.Errorf("Payment customer sign Error: ", msgobj3.ErrorMsg.Value()) // 返回错误
	}
	// Judge signature address
	e = bills.ChainPayment.FillSignByPosition(msgobj3.Sign)
	if e != nil {
		return fmt.Errorf("Fill need sign Error: ", e.Error()) // 返回错误
	}

	// 4. verify all signatures and send all signatures to the payee
	e = bills.ChainPayment.CheckMustAddressAndSigns()
	if e != nil {
		return fmt.Errorf("Check must signs Error: ", e.Error()) // 返回错误
	}
	// Send all signatures to payee
	signlist := fields.CreateEmptySignListMax255()
	signlist.Append(msgobj3.Sign)
	for _, v := range nodesigns.Signs {
		signlist.Append(v) // Local node signature
	}
	msg4 := &protocol.MsgSendChannelPayCompletedSignaturesToDownstream{
		OperationNum: operationnumber,
		AllSigns:     *signlist,
	}
	e = protocol.SendMsg(receiveCustomer.ChannelSide.wsConn, msg4)
	if e != nil {
		return fmt.Errorf("SendChannelPayCompletedSignedBillToDownstream Error: ", e.Error()) // 返回错误
	}

	// 5. save the payment bill, clear the exclusive status of the channel, and the payment is completed
	e = receiveCustomer.ChannelSide.UncheckSignSaveChannelPayReconciliationBalanceBill(bills)
	if e != nil {
		return fmt.Errorf("Receive Customer SaveChannelPayReconciliationBalanceBill Error: ", e.Error()) // 返回错误
	}
	e = payuser.ChannelSide.UncheckSignSaveChannelPayReconciliationBalanceBill(bills)
	if e != nil {
		return fmt.Errorf("Payment Customer SaveChannelPayReconciliationBalanceBill Error: ", e.Error()) // 返回错误
	}

	// Complete all payment actions
	return nil
}

*/

/*


// Start remote payment
func (s *Servicer) launchRemotePay(localnode *payroutes.PayRelayNode, newcur *chanpay.Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {

	// Expand the routing node and check whether the path is valid
	nids := msg.TargetPath.NodeIdPath
	nlen := len(nids)
	if nlen < 2 {
		return fmt.Errorf("Node id path cannot less than 2.")
	}

	// The first node must be itself
	if nids[0] != localnode.ID {
		return fmt.Errorf("First node id need %d but got %d.",
			localnode.ID, nids[0])
	}

	// open
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

	// Initiate connections to downstream nodes
	nextnode := pathnodes[1]
	urlpto := "wss"
	if s.config.DebugTest {
		urlpto = "ws" // testing environment
	}
	wsurl := fmt.Sprintf("%s://%s/relaypay/connect", urlpto, nextnode.Gateway1)

	// Connect and send initiate payment message
	var msg1 = &protocol.MsgRequestLaunchRemoteChannelPayment{}
	msg1.CopyFromInitiatePayment(msg)
	wsconn, msgobjres, _, e := protocol.OpenConnectAndSendMsgForResponseTimeout(wsurl, msg1, 15)
	if e != nil {
		return e
	}

	return nil
}

*/
