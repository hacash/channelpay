package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/core/fields"
	"sort"
)

// Add customers to the connection management pool and return to the old
func (s *Servicer) AddCustomerToPool(newcur *chanpay.Customer) error {
	if newcur.RegisteredID == 0 {
		return fmt.Errorf("Customer unregistered")
	}
	// Concurrent lock
	s.customerChgLock.Lock()
	defer s.customerChgLock.Unlock()
	//
	// Insert new
	s.customers[newcur.RegisteredID] = newcur
	// Successfully added
	return nil
}

// Remove from management pool
func (s *Servicer) RemoveCustomerFromPool(cur *chanpay.Customer) {
	// Concurrent lock
	s.customerChgLock.Lock()
	defer s.customerChgLock.Unlock()
	// remove
	s.RemoveCustomerFromPoolUnsafe(cur)
	// ok
	return
}

// Remove from management pool
func (s *Servicer) RemoveCustomerFromPoolUnsafe(cur *chanpay.Customer) {
	if cur.RegisteredID == 0 {
		return
	}
	// remove
	delete(s.customers, cur.RegisteredID)
	// ok
	return
}

// Query client connections
func (s *Servicer) FindCustomersByChannel(cid fields.ChannelId) *chanpay.Customer {

	// Concurrent lock
	s.customerChgLock.RLock()
	defer s.customerChgLock.RUnlock()

	// search
	for _, v := range s.customers {
		if v.ChannelSide.ChannelId.Equal(cid) {
			return v
		}
	}
	return nil
}

// Query client connections
func (s *Servicer) FindCustomersByAddress(addr fields.Address) []*chanpay.Customer {

	// Concurrent lock
	s.customerChgLock.RLock()
	defer s.customerChgLock.RUnlock()

	users := make([]*chanpay.Customer, 0)
	// search
	for _, v := range s.customers {
		if v.ChannelSide.RemoteAddress.Equal(addr) {
			users = append(users, v)
		}
	}
	return users
}

// Find the client connection with the largest channel capacity
// Query client connections
func (s *Servicer) FindAndStartBusinessExclusiveWithOneCustomersByAddress(addr fields.Address, payamt *fields.Amount) (*chanpay.Customer, error) {

	users := s.FindCustomersByAddress(addr)
	if len(users) == 0 {
		return nil, fmt.Errorf("Target Address Not online.") // Address not online
	}

	// Sort by collection channel capacity
	list := chanpay.CreateChannelSideConnWrapForCustomer(users)
	sort.Sort(list) // sort

	// Query and lock
	capok := false
	for i, v := range list {
		usr := users[i]
		capamt := v.GetChannelCapacityAmountForRemoteCollect()
		if !capamt.LessThan(payamt) {
			capok = true
			if usr.StartBusinessExclusive() {
				// Locking succeeded
				return usr, nil
			}
		}
	}

	if !capok {
		// Insufficient channel capacity
		return nil, fmt.Errorf("Target Address collection channel capacity not enough.")
	}

	// All channels are occupied
	return nil, fmt.Errorf("Target Address collection channels are being occupied.")
}
