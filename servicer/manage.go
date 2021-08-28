package servicer

import "fmt"

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
	pkey := string(newcur.channelId)
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
	pkey := string(cur.channelId)
	delete(s.customers, pkey)
	// ok
	return
}
