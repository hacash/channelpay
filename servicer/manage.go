package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/core/fields"
	"sort"
)

// 添加客户到连接管理池，返回旧的
func (s *Servicer) AddCustomerToPool(newcur *chanpay.Customer) error {
	if newcur.RegisteredID == 0 {
		return fmt.Errorf("Customer unregistered")
	}
	// 并发锁
	s.customerChgLock.Lock()
	defer s.customerChgLock.Unlock()
	//
	// 插入新的
	s.customers[newcur.RegisteredID] = newcur
	// 添加成功
	return nil
}

// 从管理池移除
func (s *Servicer) RemoveCustomerFromPool(cur *chanpay.Customer) {
	if cur.RegisteredID == 0 {
		return
	}
	// 并发锁
	s.customerChgLock.Lock()
	defer s.customerChgLock.Unlock()
	// 移除
	delete(s.customers, cur.RegisteredID)
	// ok
	return
}

// 查询客户端连接
func (s *Servicer) FindCustomersByChannel(cid fields.ChannelId) *chanpay.Customer {

	// 并发锁
	s.customerChgLock.RLock()
	defer s.customerChgLock.RUnlock()

	// 搜索
	for _, v := range s.customers {
		if v.ChannelSide.ChannelId.Equal(cid) {
			return v
		}
	}
	return nil
}

// 查询客户端连接
func (s *Servicer) FindCustomersByAddress(addr fields.Address) []*chanpay.Customer {

	// 并发锁
	s.customerChgLock.RLock()
	defer s.customerChgLock.RUnlock()

	users := make([]*chanpay.Customer, 0)
	// 搜索
	for _, v := range s.customers {
		if v.ChannelSide.RemoteAddress.Equal(addr) {
			users = append(users, v)
		}
	}
	return users
}

// 找出通道容量最大的客户端连接
// 查询客户端连接
func (s *Servicer) FindAndStartBusinessExclusiveWithOneCustomersByAddress(addr fields.Address, payamt *fields.Amount) (*chanpay.Customer, error) {

	users := s.FindCustomersByAddress(addr)
	if len(users) == 0 {
		return nil, fmt.Errorf("Target Address Not online.") // 地址不在线
	}

	// 按收款通道容量排序
	list := chanpay.CreateChannelSideConnWrapForCustomer(users)
	sort.Sort(list) // 排序

	// 查询并锁定
	capok := false
	for i, v := range list {
		usr := users[i]
		capamt := v.GetChannelCapacityAmountForRemoteCollect()
		if !capamt.LessThan(payamt) {
			capok = true
			if usr.StartBusinessExclusive() {
				// 锁定成功
				return usr, nil
			}
		}
	}

	if !capok {
		// 通道容量不足
		return nil, fmt.Errorf("Target Address collection channel capacity not enough.")
	}

	// 通道全部被占用
	return nil, fmt.Errorf("Target Address collection channels are being occupied.")
}
