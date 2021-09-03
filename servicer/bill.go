package servicer

import (
	"fmt"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"time"
)

/**
 * 票据相关
 */

// 创建对账单
// cusispay 客户方支付还是收款
func (s *Servicer) CreateChannelPayTransferProveBody(usr *Customer, trsAmt *fields.Amount, cusispay bool) (*channel.ChannelChainTransferProveBodyInfo, error) {
	// 创建支付方对账单
	var autoNumber1 = fields.VarUint8(1)
	if usr.latestReconciliationBalanceBill != nil {
		autoNumber1 = fields.VarUint8(usr.latestReconciliationBalanceBill.ChannelAutoNumber() + 1)
	}
	trsbody := &channel.ChannelChainTransferProveBodyInfo{
		ChannelId:           usr.channelId,
		Mode:                channel.ChannelTransferProveBodyPayModeNormal,
		LeftAmount:          fields.Amount{},
		RightAmount:         fields.Amount{},
		ChannelReuseVersion: usr.channelInfo.ReuseVersion,
		BillAutoNumber:      autoNumber1,
	}
	paySideBls := usr.GetChannelCapacityAmountForPay()
	collectSideBls := usr.GetChannelCapacityAmountForCollect()
	if cusispay == false {
		// 客户端收款, 调换位置
		paySideBls, collectSideBls = collectSideBls, paySideBls
	}
	// 检查
	if paySideBls.LessThan(trsAmt) {
		return nil, fmt.Errorf("Channel customer balance not enough.")
	}
	paySideDis, e := paySideBls.Sub(trsAmt)
	if e != nil {
		return nil, fmt.Errorf("Channel balance distribution error: %s", e.Error())
	}
	collectSideDis, e := collectSideBls.Add(trsAmt)
	if e != nil {
		return nil, fmt.Errorf("Channel balance distribution error: %s", e.Error())
	}
	// 通道分配
	cusisleft := usr.CustomerAddressIsLeft()
	if (cusisleft && cusispay) || (!cusisleft && !cusispay) {
		trsbody.LeftAmount, trsbody.RightAmount = *paySideDis, *collectSideDis
	} else {
		trsbody.LeftAmount, trsbody.RightAmount = *collectSideDis, *paySideDis
	}
	// 创建成功
	return trsbody, nil
}

// 创建通道支付票据
func (s *Servicer) CreateChannelPayTransferTransactionForLocalPay(payusr *Customer, collectusr *Customer, trsAmt *fields.Amount, realpayamtwithfee *fields.Amount, orderHashCheck fields.HashHalfChecker) (*channel.ChannelPayBillAssemble, error) {

	// 创建支付方对账单
	ispay1 := true
	paybody, e := s.CreateChannelPayTransferProveBody(payusr, realpayamtwithfee, ispay1)
	if e != nil {
		return nil, fmt.Errorf("CreateChannelPayTransferProveBody for Pay side Error: %s", e.Error())
	}

	// 创建收款方对账单
	ispay2 := false
	collectbody, e := s.CreateChannelPayTransferProveBody(payusr, trsAmt, ispay2)
	if e != nil {
		return nil, fmt.Errorf("CreateChannelPayTransferProveBody for Collect side Error: %s", e.Error())
	}

	// realpayamtwithfee 和 trsAmt 的差额即为节点的手续费了

	// 建立票据集合
	proveBodys := make([]*channel.ChannelChainTransferProveBodyInfo, 2)
	proveBodys[0] = paybody
	proveBodys[1] = collectbody

	// 建立支付签名票据
	bdchecker1 := paybody.SignStuffHashHalfChecker()
	bdchecker2 := collectbody.SignStuffHashHalfChecker()
	curtimes := fields.BlockTxTimestamp(time.Now().Unix())
	billforsign := &channel.OffChainFormPaymentChannelTransfer{
		Timestamp:                curtimes,
		OrderNoteHashHalfChecker: orderHashCheck,
		MustSignCount:            0,
		MustSignAddresses:        nil,
		ChannelCount:             2,
		ChannelTransferProveHashHalfCheckers: []fields.HashHalfChecker{
			bdchecker1, bdchecker2,
		},
		MustSigns: nil,
	}
	// 地址去重和排序
	billforsign.MustSignCount, billforsign.MustSignAddresses =
		channel.CleanSortMustSignAddresses([]fields.Address{
			payusr.channelInfo.LeftAddress, payusr.channelInfo.RightAddress,
			collectusr.channelInfo.LeftAddress, collectusr.channelInfo.RightAddress,
		})

	// 空签名填充
	signnum := int(billforsign.MustSignCount)
	billforsign.MustSigns = make([]fields.Sign, signnum)
	for i := 0; i < signnum; i++ {
		billforsign.MustSigns[i] = fields.CreateEmptySign()
	}

	// 创建完毕，返回
	allbill := &channel.ChannelPayBillAssemble{
		ProveBodys: &channel.ChannelPayProveBodyList{
			Count: fields.VarUint1(2),
			ProveBodys: []*channel.ChannelChainTransferProveBodyInfo{
				paybody, collectbody,
			},
		},
		ChainPayment: billforsign,
	}

	return allbill, nil
}
