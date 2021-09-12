package servicer

import (
	"fmt"
	"github.com/hacash/core/fields"
	"sort"
)

// 添加客户到连接管理池，返回旧的
func (s *Servicer) AddCustomerToPool(newcur *Customer) (*Customer, error) {
	if newcur.IsRegistered == false {
		return nil, fmt.Errorf("Customer unregistered")
	}
	// 并发锁
	s.customerChgLock.Lock()
	defer s.customerChgLock.Unlock()
	//
	var oldcur *Customer = nil
	// 查找旧的
	pkey := string(newcur.ChannelSide.channelId)
	if old, hav := s.customers[pkey]; hav {
		oldcur = old
	}
	// 插入新的
	s.customers[pkey] = newcur
	// 添加成功
	return oldcur, nil
}

// 从管理池移除
func (s *Servicer) RemoveCustomerFromPool(cur *Customer) {
	if cur.IsRegistered == false {
		return
	}
	// 并发锁
	s.customerChgLock.Lock()
	defer s.customerChgLock.Unlock()
	// 移除
	pkey := string(cur.ChannelSide.channelId)
	delete(s.customers, pkey)
	// ok
	return
}

// 查询客户端连接
func (s *Servicer) FindCustomersByAddress(addr fields.Address) []*Customer {

	// 并发锁
	s.customerChgLock.RLock()
	defer s.customerChgLock.RUnlock()

	users := make([]*Customer, 0)
	// 搜索
	for _, v := range s.customers {
		if v.ChannelSide.remoteAddress.Equal(addr) {
			users = append(users, v)
		}
	}
	return users
}

// 找出通道容量最大的客户端连接
// 查询客户端连接
func (s *Servicer) FindAndStartBusinessExclusiveWithOneCustomersByAddress(addr fields.Address, payamt *fields.Amount) (*Customer, error) {

	users := s.FindCustomersByAddress(addr)
	if len(users) == 0 {
		return nil, fmt.Errorf("Target Address Not online.") // 地址不在线
	}

	// 按收款通道容量排序
	list := CreateChannelSideConnWrapForCustomer(users)
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
