package servicer

import "fmt"

// 全新登录一个客户端
func (s *Servicer) LoginNewCustomer(newcur *Customer) error {

	// 读取最新对账单
	bill, e := s.billstore.GetLastestBalanceBill(newcur.channelId)
	if e != nil {
		return e
	}
	newcur.latestReconciliationBalanceBill = bill

	// 账户
	cusAddr := newcur.customerAddress
	customerIsLeft := cusAddr.Equal(bill.LeftAddress())
	customerIsRight := cusAddr.Equal(bill.RightAddress())
	if !customerIsLeft && !customerIsRight {
		return fmt.Errorf("address %s is not belong to channel %s",
			cusAddr.ToReadable(),
			newcur.channelId.ToHex())
	}
	if customerIsLeft {
		newcur.servicerAddress = bill.RightAddress()
	} else {
		newcur.servicerAddress = bill.LeftAddress()
	}

	// 完成登录
	return nil
}
