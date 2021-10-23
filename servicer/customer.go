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
	s.customerChgLock.RLock()
	for _, v := range s.customers {
		susary = append(susary, v)
	}
	s.customerChgLock.RUnlock()

	// 检查时间，30秒心跳过期
	tnck := time.Now().Unix() - 30
	for _, v := range susary {
		if v.GetLastestHeartbeatTime().Unix() < tnck {
			// 超过30秒没有心跳，断开连接
			v.ChannelSide.WsConn.Close()
		}
	}
}
