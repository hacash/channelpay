package servicer

import (
	"github.com/hacash/channelpay/chanpay"
	"time"
)

/**
 * 客户端活跃检测
 */
func (s *Servicer) checkCustomerActive() {

	var susary = make([]*chanpay.Customer, 0)
	s.customerChgLock.Lock()
	for _, v := range s.customers {
		susary = append(susary, v)
	}

	// Check time, 30 seconds heartbeat expired
	tnck := time.Now().Unix() - 30
	for _, v := range susary {
		if v.GetLastestHeartbeatTime().Unix() < tnck {
			// No heartbeat for more than 30 seconds, disconnect
			//fmt.Println("v.GetLastestHeartbeatTime().Unix() < tnck CLOSE")
			v.ChannelSide.WsConn.Close()
			s.RemoveCustomerFromPoolUnsafe(v)
		}
	}
	s.customerChgLock.Unlock()
}
