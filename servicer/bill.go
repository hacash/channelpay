package servicer

/**
 * 票据相关
 *

// Create statement
// Cusispay customer payment or collection
func (s *Servicer) CreateChannelPayTransferProveBody(usr *chanpay.Customer, trsAmt *fields.Amount, cusispay bool) (*channel.ChannelChainTransferProveBodyInfo, error) {
	// Create payer statement
	var autoNumber1 = fields.VarUint8(1)
	usrbill := usr.ChannelSide.GetReconciliationBill()
	if usrbill != nil {
		autoNumber1 = fields.VarUint8(usrbill.ChannelAutoNumber() + 1)
	}
	trsbody := &channel.ChannelChainTransferProveBodyInfo{
		GetChannelId:      usr.ChannelSide.channelId,
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
		// Client collection, exchange location
		paySideBls, collectSideBls = collectSideBls, paySideBls
	}
	// inspect
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
	// Payment direction, channel allocation
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
	// Created successfully
	return trsbody, nil
}

// Create channel payment ticket
func (s *Servicer) CreateChannelPayTransferTransactionForLocalPay(payusr *chanpay.Customer, collectusr *chanpay.Customer, trsAmt *fields.Amount, realpayamtwithfee *fields.Amount, orderHashCheck fields.HashHalfChecker) (*channel.ChannelPayCompleteDocuments, error) {

	// Create payer statement
	ispay1 := true
	paybody, e := s.CreateChannelPayTransferProveBody(payusr, realpayamtwithfee, ispay1)
	if e != nil {
		return nil, fmt.Errorf("CreateChannelPayTransferProveBody for Pay side Error: %s", e.Error())
	}

	// Create payee statement
	ispay2 := false
	collectbody, e := s.CreateChannelPayTransferProveBody(payusr, trsAmt, ispay2)
	if e != nil {
		return nil, fmt.Errorf("CreateChannelPayTransferProveBody for Collect side Error: %s", e.Error())
	}

	// The difference between realpayamtwithfee and trsamt is the service charge of the node

	// Create ticket collection
	proveBodys := make([]*channel.ChannelChainTransferProveBodyInfo, 2)
	proveBodys[0] = paybody
	proveBodys[1] = collectbody

	// Create payment signature bill
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
	// Address de duplication and sorting
	billforsign.MustSignCount, billforsign.MustSignAddresses =
		channel.CleanSortMustSignAddresses([]fields.Address{
			payusr.ChannelSide.channelInfo.LeftAddress, payusr.ChannelSide.channelInfo.RightAddress,
			collectusr.ChannelSide.channelInfo.LeftAddress, collectusr.ChannelSide.channelInfo.RightAddress,
		})

	// Empty signature fill
	signnum := int(billforsign.MustSignCount)
	billforsign.MustSigns = make([]fields.Sign, signnum)
	for i := 0; i < signnum; i++ {
		billforsign.MustSigns[i] = fields.CreateEmptySign()
	}

	// After creation, return to
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



*/
