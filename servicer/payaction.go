package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/payroutes"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/channelpay/servicer/datasources"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
	"sync"
)

/**
 * 单次支付行为操作
 */
type ChannelPayActionInstance struct {

	// 状态锁
	statusUpdateMux sync.Mutex

	// 本地服务节点
	localServicerNode *payroutes.PayRelayNode

	// 通道链上游、下游连接端
	upstreamSide   *RelayPaySettleNoder // 通道上游
	downstreamSide *RelayPaySettleNoder // 通道下游

	// 或者为支付端、收款端
	payCustomer     *Customer // 支付端客户端
	collectCustomer *Customer // 收款端客户端

	// 通道长度
	channelLength         int  // 通道路径长度
	ourProveBodyCompleted bool // 我们自己是否已经完成签名

	// 目标交易票据
	billDocuments            *channel.ChannelPayCompleteDocuments
	transactionDistinguishId fields.VarUint8
	allSignsCompleted        bool

	// 签名机器
	signMachine datasources.DataSourceOfSignatureMachine

	// 签名地址
	ourMustSignAddresses []fields.Address
	ourSignCompleted     bool // 我们自己是否已经完成签名

}

// 新创建
func NewChannelPayActionInstance() *ChannelPayActionInstance {
	return &ChannelPayActionInstance{
		channelLength:            0,
		localServicerNode:        nil,
		allSignsCompleted:        false,
		transactionDistinguishId: 0,
		signMachine:              nil,
		ourMustSignAddresses:     nil, // 必须签名地址
		ourSignCompleted:         false,
		ourProveBodyCompleted:    false,
	}
}

func (c *ChannelPayActionInstance) checkOurAddress(addr fields.Address) bool {
	if c.ourMustSignAddresses != nil {
		for _, v := range c.ourMustSignAddresses {
			if v.Equal(addr) {
				return true
			}
		}
	}
	return false
}

// 设定签名机器
func (c *ChannelPayActionInstance) SetSignatureMachine(machine datasources.DataSourceOfSignatureMachine) {
	c.signMachine = machine
}

// 设定签名地址
func (c *ChannelPayActionInstance) SetMustSignAddresses(addrs []fields.Address) {
	c.ourMustSignAddresses = addrs
}

// 检查是否可以签名，可以的话直接完成签名
func (c *ChannelPayActionInstance) CheckMaybeCanDoSign() (bool, error) {

	// 我是否已经全部签名
	if c.ourSignCompleted {
		return true, nil // 仅签名一次
	}

	// 检查对账单完成度
	var mustSignList = make([]fields.Address, 0)
	var prevlast fields.Address = nil
	for i, v := range c.billDocuments.ProveBodys.ProveBodys {
		chr := c.billDocuments.ChainPayment.ChannelTransferProveHashHalfCheckers[i]
		if v == nil || chr == nil {
			return false, fmt.Errorf("ProveBodys not complete.")
		}
		// 检查
		var a1, a2 = v.LeftAddress, v.RightAddress
		if v.PayDirection == 2 { // 如果从又往左支付
			a2, a1 = a1, a2
		}
		if prevlast.Equal(a1) {
			mustSignList = append(mustSignList, a2)
		} else {
			mustSignList = append(mustSignList, a1, a2)
		}
		prevlast = a2 // 记录
	}

	// 检查签名
	for i := len(mustSignList) - 1; i >= 0; i-- {
		one := mustSignList[i]
		if c.checkOurAddress(one) {
			break // 检查完毕
		}
		// 检查下游签名是否完成
		cke := c.billDocuments.ChainPayment.CheckOneAddressSign(one)
		if cke != nil {
			// 下游签名检查失败
			return false, nil // 不返回错误，等待下次再检查
		}
	}

	// 执行我的签名
	mysigns, e := c.signMachine.CheckPaydocumentAndFillNeedSignature(c.billDocuments, c.ourMustSignAddresses)
	if e != nil {
		// 返回签名错误
		return false, fmt.Errorf("signature fail: %s", e.Error())
	}

	// 广播我的签名
	sgmsg := &protocol.MsgBroadcastChannelStatementSignature{
		TransactionDistinguishId: c.transactionDistinguishId,
		Signs:                    *mysigns,
	}
	// 广播消息
	c.BroadcastMessage(sgmsg)

	// 填充我的签名
	for _, v := range mysigns.Signs {
		e := c.billDocuments.ChainPayment.FillSignByPosition(v)
		if e != nil {
			// 返回填充签名失败
			return false, fmt.Errorf("fill signature fail: %s", e.Error())
		}
	}

	// 我签名成功
	c.ourSignCompleted = true // 标记
	return true, nil
}

// 创建对账单
func (c *ChannelPayActionInstance) createMyProveBodyByRemotePay(collectAmt *fields.Amount) (*channel.ChannelChainTransferProveBodyInfo, error) {

	// 找出通道
	var chanSide *ChannelSideConn = nil
	if c.upstreamSide != nil {
		chanSide = c.upstreamSide.ChannelSide
	} else if c.payCustomer != nil {
		chanSide = c.payCustomer.ChannelSide
	} else {
		return nil, fmt.Errorf("not find up side ChannelSide ptr.")
	}

	// 检查容量
	amtcap := chanSide.GetChannelCapacityAmountOfRemote()
	if amtcap.LessThan(collectAmt) {
		return nil, fmt.Errorf("up side channel capacity balance not enough.")
	}

	// 创建对账单
	chaninfo := chanSide.channelInfo
	reuseversion := uint32(chaninfo.ReuseVersion)
	billautonumber := uint64(1)
	lastbill := chanSide.latestReconciliationBalanceBill
	oldleftamt := chaninfo.LeftAmount
	oldrightamt := chaninfo.RightAmount
	if lastbill != nil {
		oldleftamt = lastbill.LeftAmount()
		oldrightamt = lastbill.RightAmount()
		blrn, blan := lastbill.ChannelReuseVersionAndAutoNumber()
		if reuseversion != blrn {
			return nil, fmt.Errorf("Channel Reuse Version %d in last bill and %d in channel info not match.", blrn, reuseversion)
		}
		billautonumber = uint64(blan) + 1 // 自增
	}
	// 方向
	paydirection := 2
	oldpayamt := oldrightamt
	oldcollectamt := oldleftamt
	remoteisleft := chanSide.RemoteAddressIsLeft()
	if remoteisleft {
		paydirection = 1                                    // 总是对方支付给我
		oldpayamt, oldcollectamt = oldcollectamt, oldpayamt // 反向
	}
	// 计算最新分配
	newpayamt, e := oldpayamt.Sub(collectAmt)
	if e != nil {
		return nil, fmt.Errorf("pay error: %s", e.Error())
	}
	if newpayamt == nil || newpayamt.IsNegative() {
		return nil, fmt.Errorf("pay error: balance not enough.")
	}
	newcollectamt, e := oldcollectamt.Sub(collectAmt)
	if e != nil {
		return nil, fmt.Errorf("collect error: %s", e.Error())
	}
	if newcollectamt == nil || newpayamt.IsNegative() {
		return nil, fmt.Errorf("collect error: balance is Negative.")
	}
	newleftamt := newcollectamt
	newrightamt := newpayamt
	if remoteisleft {
		newleftamt, newrightamt = newrightamt, newleftamt // 反向
	}
	// 创建
	body := &channel.ChannelChainTransferProveBodyInfo{
		ChannelId:      chanSide.channelId,
		ReuseVersion:   fields.VarUint4(reuseversion),
		BillAutoNumber: fields.VarUint8(billautonumber),
		PayDirection:   fields.VarUint1(paydirection),
		PayAmount:      *collectAmt,
		LeftAddress:    chaninfo.LeftAddress,
		RightAddress:   chaninfo.RightAddress,
		LeftBalance:    *newleftamt,
		RightBalance:   *newrightamt,
	}

	// 返回
	return body, nil
}

// 是否要报告我的交易体
func (c *ChannelPayActionInstance) CheckMaybeReportMyProveBody(msg *protocol.MsgRequestInitiatePayment) (bool, error) {
	if c.ourProveBodyCompleted {
		return true, nil // 防止重复报告
	}

	var e error
	var bodyindex int = 0
	var bodyinfo *channel.ChannelChainTransferProveBodyInfo = nil

	// 作为最终收款方首次报告
	if msg != nil && c.downstreamSide == nil || c.collectCustomer == nil {
		upside := c.upstreamSide
		if upside == nil {
			return false, fmt.Errorf("upstreamSide os nil")
		}
		// 创建对账单
		bodyindex = c.channelLength - 1
		bodyinfo, e = c.createMyProveBodyByRemotePay(&msg.PayAmount)
		if e != nil {
			return false, fmt.Errorf("create my prove body by remote pay error: %s", e.Error())
		}

	} else {

		// 检查下游是否已经填充报告
		var downSide *ChannelSideConn = nil
		if c.downstreamSide != nil {
			downSide = c.downstreamSide.ChannelSide
		} else if c.collectCustomer != nil {
			downSide = c.collectCustomer.ChannelSide
		} else {
			return false, fmt.Errorf("channel pay down stream side not find")
		}
		// 下游支付
		var downSideColletAmt *fields.Amount = nil
		var payaddr *fields.Address = nil
		bdlist := c.billDocuments.ProveBodys.ProveBodys
		for i := len(bdlist) - 1; i >= 0; i-- {
			one := bdlist[i]
			if one.ChannelId.Equal(downSide.channelId) {
				bodyindex = i - 1
				downSideColletAmt = &one.PayAmount
				payaddr = &one.LeftAddress
				if one.PayDirection == 2 {
					payaddr = &one.RightAddress
				}
			}
		}
		if downSideColletAmt == nil {
			// 下游还没有广播对账单，等待下一次检测
			// 不返回错误
			return false, nil
		}
		// 检查付款段是否是我
		if payaddr.NotEqual(downSide.ourAddress) {
			return false, fmt.Errorf("pay address is not our address.")
		}
		// 检测我是否为付款端
		var upSide *ChannelSideConn = nil
		if c.upstreamSide != nil {
			upSide = c.upstreamSide.ChannelSide
		} else if c.payCustomer != nil {
			upSide = c.payCustomer.ChannelSide
		} else {
			// 没有上游，我就是付款端，不用广播什么，等待签名即可
			return false, nil
		}
		if bodyindex < 0 {
			// 没有上游，我就是付款端，不用广播什么，等待签名即可
			return false, nil
		}
		// 加上手续费
		// 加上我自己节点的手续费
		lcsnode := c.localServicerNode
		if lcsnode == nil {
			return false, fmt.Errorf("local servicer node is nil.")
		}
		// 手续费
		appendfee := lcsnode.PredictFeeForPay(downSideColletAmt)
		newpayamt, e := downSideColletAmt.Add(appendfee)
		if e != nil {
			return false, fmt.Errorf("add predict fee for pay error: %s", e.Error())
		}
		// 创建对账单
		bodyinfo, e = c.createMyProveBodyByRemotePay(newpayamt)
		if e != nil {
			return false, fmt.Errorf("create my prove body of channel id %s by remote pay error: %s", upSide.channelId.ToHex(), e.Error())
		}
	}

	// 是否广播对账单
	if bodyinfo == nil {
		return false, nil
	}
	// 广播
	msgbd := &protocol.MsgBroadcastChannelStatementProveBody{
		TransactionDistinguishId: c.transactionDistinguishId,
		ProveBodyIndex:           fields.VarUint1(bodyindex),
		ProveBodyInfo:            bodyinfo,
	}
	c.BroadcastMessage(msgbd)      // 广播消息
	c.ourProveBodyCompleted = true // 报告成功
	return true, nil

}

// 创建交易票据，通过消息
func (c *ChannelPayActionInstance) InitCreateEmptyBillDocumentsByInitPayMsg(msg *protocol.MsgRequestInitiatePayment) error {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	// 通道长度
	var paychanlen = int(msg.TargetPath.NodeIdCount) + 1
	c.channelLength = paychanlen
	c.allSignsCompleted = false
	c.transactionDistinguishId = msg.TransactionDistinguishId
	// 交易票据
	bodys := make([]*channel.ChannelChainTransferProveBodyInfo, paychanlen)
	for i := 0; i < paychanlen; i++ {
		bodys[i] = nil
	}
	c.billDocuments = &channel.ChannelPayCompleteDocuments{
		ProveBodys: &channel.ChannelPayProveBodyList{
			Count:      fields.VarUint1(paychanlen),
			ProveBodys: bodys,
		},
		ChainPayment: &channel.OffChainFormPaymentChannelTransfer{
			Timestamp:                            msg.Timestamp,
			OrderNoteHashHalfChecker:             msg.OrderNoteHashHalfChecker,
			MustSignCount:                        0,
			MustSignAddresses:                    make([]fields.Address, 0),
			ChannelCount:                         fields.VarUint1(paychanlen),
			ChannelTransferProveHashHalfCheckers: make([]fields.HashHalfChecker, paychanlen),
			MustSigns:                            make([]fields.Sign, 0),
		},
	}

	// 如果我是最终收款方，我需要第一个报告我的交易体
	_, e := c.CheckMaybeReportMyProveBody(msg)
	return e
}

// 全部完成，销毁所有资源
func (c *ChannelPayActionInstance) Destroy() {
	if c.upstreamSide != nil {
		c.upstreamSide.ChannelSide.wsConn.Close()
		c.upstreamSide.ClearBusinessExclusive() // 解除独占
	}
	if c.downstreamSide != nil {
		c.downstreamSide.ChannelSide.wsConn.Close()
		c.downstreamSide.ClearBusinessExclusive() // 解除独占
	}
	if c.payCustomer != nil {
		c.payCustomer.ClearBusinessExclusive() // 解除独占
	}
	if c.collectCustomer != nil {
		c.collectCustomer.ClearBusinessExclusive() // 解除独占
	}
}

// 获取另一边连接，用于转发消息
func (c *ChannelPayActionInstance) GetUpOrDownStreamNegativeDirection(upOrDownStream bool) *websocket.Conn {
	var conn *websocket.Conn = nil
	if upOrDownStream {
		if c.downstreamSide != nil {
			conn = c.downstreamSide.ChannelSide.wsConn
		}
		if c.collectCustomer != nil {
			conn = c.collectCustomer.ChannelSide.wsConn
		}
	} else {
		if c.upstreamSide != nil {
			conn = c.upstreamSide.ChannelSide.wsConn
		}
		if c.payCustomer != nil {
			conn = c.payCustomer.ChannelSide.wsConn
		}
	}
	return conn
}

// 广播消息
func (c *ChannelPayActionInstance) BroadcastMessage(msg protocol.Message) {
	wsconn1 := c.GetUpOrDownStreamNegativeDirection(true)
	wsconn2 := c.GetUpOrDownStreamNegativeDirection(false)
	if wsconn1 != nil {
		protocol.SendMsg(wsconn1, msg)
	}
	if wsconn2 != nil {
		protocol.SendMsg(wsconn2, msg)
	}
}

// 启动某一边消息监听
func (c *ChannelPayActionInstance) StartOneSideMessageListen(upOrDownStream bool, conn *websocket.Conn) {
	go func() {
		for {
			// 读取消息
			msgobj, _, e := protocol.ReceiveMsg(conn)
			if e != nil {
				break // 错误
			}
			if msgobj != nil {
				var e error = nil
				switch msgobj.Type() {
				// 错误到达
				case protocol.MsgTypeBroadcastChannelStatementError:
					c.ChannelPayErrorArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementError))
				// 对账单到达
				case protocol.MsgTypeBroadcastChannelStatementProveBody:
					e = c.ChannelPayProveBodyArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementProveBody))
				// 签名到达
				case protocol.MsgTypeBroadcastChannelStatementSignature:
					e = c.ChannelPaySignatureArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementSignature))
				// 支付成功
				case protocol.MsgTypeBroadcastChannelStatementSuccessed:
					c.ChannelPaySuccessArrive(upOrDownStream, msgobj.(*protocol.MsgBroadcastChannelStatementSuccessed))
				}
				// 错误处理
				if e != nil {
					// 广播错误
					msg := &protocol.MsgBroadcastChannelStatementError{
						ErrCode: 1,
						ErrTip:  fields.CreateStringMax65535(e.Error()),
					}
					c.BroadcastMessage(msg)
				}
			}
		}
		// 发生错误断开连接

	}()
}

// 填充交易票据，对账单到达
func (c *ChannelPayActionInstance) ChannelPayProveBodyArrive(upOrDownStream bool, msg *protocol.MsgBroadcastChannelStatementProveBody) error {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	// 检查body
	bodychecker := msg.ProveBodyInfo.GetSignStuffHashHalfChecker()
	// 填入
	checkerlist := c.billDocuments.ChainPayment.ChannelTransferProveHashHalfCheckers
	cid := int(msg.ProveBodyIndex)
	if len(checkerlist) < cid {
		return fmt.Errorf("ProveBodyIndex overflow pay channel chain length.")
	}
	checkerlist[cid] = bodychecker                                 // 填充
	c.billDocuments.ProveBodys.ProveBodys[cid] = msg.ProveBodyInfo // 填充
	// 添加地址
	addrlist := make([]fields.Address, 2)
	addrlist[0] = msg.ProveBodyInfo.GetLeftAddress()
	addrlist[1] = msg.ProveBodyInfo.GetRightAddress()
	addrlen, addrs := fields.CleanAddressListByCharacterSort(c.billDocuments.ChainPayment.MustSignAddresses, addrlist)
	// 填充票据
	c.billDocuments.ChainPayment.MustSignCount = fields.VarUint1(addrlen)
	c.billDocuments.ChainPayment.MustSignAddresses = addrs
	allsigns := make([]fields.Sign, addrlen)
	for i := 0; i < int(addrlen); i++ {
		allsigns[i] = fields.CreateEmptySign()
	}
	c.billDocuments.ChainPayment.MustSigns = allsigns

	// 向另一方转发交易体
	otherside := c.GetUpOrDownStreamNegativeDirection(upOrDownStream)
	if otherside != nil {
		// 转发
		protocol.SendMsg(otherside, msg)
	}

	// 是否要报告我的交易体
	_, e := c.CheckMaybeReportMyProveBody(nil)
	if e != nil {
		return e
	}

	// 是否可以签名
	_, e = c.CheckMaybeCanDoSign()
	if e != nil {
		return e
	}

	// ok
	return nil

}

// 签名数据到达
func (c *ChannelPayActionInstance) ChannelPaySignatureArrive(upOrDownStream bool, msg *protocol.MsgBroadcastChannelStatementSignature) error {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	// 填充签名，不做检查，忽略错误
	for _, v := range msg.Signs.Signs {
		e := c.billDocuments.ChainPayment.FillSignByPosition(v)
		if e != nil {
			return e
		}
	}

	// 向另一方转发签名
	otherside := c.GetUpOrDownStreamNegativeDirection(upOrDownStream)
	if otherside != nil {
		// 转发
		protocol.SendMsg(otherside, msg)
	}

	// 是否可以签名
	_, e := c.CheckMaybeCanDoSign()
	if e != nil {
		return e
	}

	// 如果我是最后收款方，则最后一个签名到达时，我负责广播支付成功完成的消息
	if c.downstreamSide == nil && c.collectCustomer == nil {
		e := c.billDocuments.ChainPayment.CheckMustAddressAndSigns()
		if e == nil {
			// 所有签名全部完成，广播完成消息
			okmsg := &protocol.MsgBroadcastChannelStatementSuccessed{
				SuccessTip: fields.CreateStringMax65535(""),
			}
			c.BroadcastMessage(okmsg)
			// 断开
			if otherside != nil {
				otherside.Close() // 断开连接
			}
		}
	}

	// ok
	return nil
}

// 错误到达
func (c *ChannelPayActionInstance) ChannelPayErrorArrive(upOrDownStream bool, msg *protocol.MsgBroadcastChannelStatementError) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	// 将错误转发给另一方
	otherside := c.GetUpOrDownStreamNegativeDirection(upOrDownStream)
	if otherside != nil {
		protocol.SendMsg(otherside, msg)
	}

	// 全部结束
	c.Destroy()
}

// 支付成功完成
func (c *ChannelPayActionInstance) ChannelPaySuccessArrive(upOrDownStream bool, msg *protocol.MsgBroadcastChannelStatementSuccessed) {
	c.statusUpdateMux.Lock()
	defer c.statusUpdateMux.Unlock()

	// 将成功转发给上游，单向消息
	otherside := c.GetUpOrDownStreamNegativeDirection(false)
	if otherside != nil {
		protocol.SendMsg(otherside, msg)
	}

	// 全部结束
	c.Destroy()
}
