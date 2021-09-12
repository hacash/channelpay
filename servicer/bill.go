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
	usrbill := usr.ChannelSide.GetReconciliationBill()
	if usrbill != nil {
		autoNumber1 = fields.VarUint8(usrbill.ChannelAutoNumber() + 1)
	}
	trsbody := &channel.ChannelChainTransferProveBodyInfo{
		ChannelId:      usr.ChannelSide.channelId,
		ReuseVersion:   usr.ChannelSide.channelInfo.ReuseVersion,
		BillAutoNumber: autoNumber1,
		PayDirection:   1, // 后续判断支付方向
		PayAmount:      *trsAmt,
		Mode:           channel.ChannelTransferProveBodyPayModeNormal,
		LeftBalance:    fields.Amount{},
		RightBalance:   fields.Amount{},
	}
	paySideBls := usr.GetChannelCapacityAmountForRemotePay()
	collectSideBls := usr.GetChannelCapacityAmountForRemoteCollect()
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
	// 支付方向，通道分配
	cusisleft := usr.CustomerAddressIsLeft()
	if cusisleft && cusispay {
	}
	if (cusisleft && cusispay) || (!cusisleft && !cusispay) {
		// left => right
		trsbody.PayDirection = channel.ChannelTransferProveBodyPayDirectionLeftToRight
		trsbody.LeftBalance, trsbody.RightBalance = *paySideDis, *collectSideDis
	} else {
		// right => left
		trsbody.PayDirection = channel.ChannelTransferProveBodyPayDirectionRightToLeft
		trsbody.LeftBalance, trsbody.RightBalance = *collectSideDis, *paySideDis
	}
	// 创建成功
	return trsbody, nil
}

// 创建通道支付票据
func (s *Servicer) CreateChannelPayTransferTransactionForLocalPay(payusr *Customer, collectusr *Customer, trsAmt *fields.Amount, realpayamtwithfee *fields.Amount, orderHashCheck fields.HashHalfChecker) (*channel.ChannelPayCompleteDocuments, error) {

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
	bdchecker1 := paybody.GetSignStuffHashHalfChecker()
	bdchecker2 := collectbody.GetSignStuffHashHalfChecker()
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
			payusr.ChannelSide.channelInfo.LeftAddress, payusr.ChannelSide.channelInfo.RightAddress,
			collectusr.ChannelSide.channelInfo.LeftAddress, collectusr.ChannelSide.channelInfo.RightAddress,
		})

	// 空签名填充
	signnum := int(billforsign.MustSignCount)
	billforsign.MustSigns = make([]fields.Sign, signnum)
	for i := 0; i < signnum; i++ {
		billforsign.MustSigns[i] = fields.CreateEmptySign()
	}

	// 创建完毕，返回
	allbill := &channel.ChannelPayCompleteDocuments{
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
