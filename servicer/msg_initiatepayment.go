package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
)

/**
 * 发起支付
 */

func (s *Servicer) MsgHandlerRequestInitiatePayment(newcur *Customer, msg *protocol.MsgRequestInitiatePayment) {

	// 返回错误消息
	errorReturn := func(e error) {
		errmsg := &protocol.MsgError{
			ErrCode: 0,
			ErrTip:  fields.CreateStringMax65535(e.Error()),
		}
		protocol.SendMsg(newcur.wsConn, errmsg)
	}

	var e error

	// 地址
	targetAddr := &protocol.ChannelAccountAddress{}
	e = targetAddr.Parse(msg.PayeeChannelAddr.Value())
	if e != nil {
		errorReturn(e)
		return
	}

	// 本地还是远程支付
	if targetAddr.CompareServiceName(s.config.SelfIdentificationName) {
		e = s.localPay(newcur, msg, targetAddr)
	} else {
		e = s.remotePay(newcur, msg, targetAddr)
	}

	// 返回错误
	if e != nil {
		errorReturn(e)
		return
	}
}

// 开始本地支付
func (s *Servicer) localPay(newcur *Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {

	// 取出收款目标地址的连接
	targetCuntomers := make([]*Customer, 0)
	s.customerChgLock.RLock()
	for _, v := range s.customers {
		if v.customerAddress.Equal(targetAddr.Address) {
			targetCuntomers = append(targetCuntomers, v)
		}
	}
	s.customerChgLock.RUnlock()

	// 是否有在线的客户端
	cusnum := len(targetCuntomers)
	if cusnum == 0 {
		return fmt.Errorf("Target address %s is offline.", targetAddr.Address.ToReadable())
	}

	// 筛选最适合收款
	var targetCuntomer = targetCuntomers[0]
	var chanwideamt = targetCuntomer.GetChannelCapacityAmountForCollect()
	if cusnum > 1 {
		// 找出收款最大的通道容量
		for i := 1; i < len(targetCuntomers); i++ {
			v := targetCuntomers[i]
			wideamt := v.GetChannelCapacityAmountForCollect()
			if wideamt.MoreThan(chanwideamt) {
				chanwideamt = wideamt
				targetCuntomer = v
			}
		}
	}

	// 检查通道容量
	if chanwideamt.LessThan(&msg.PayAmount) {
		// 通道收款容量不足
		return fmt.Errorf("Target address channel collect capacity %s insufficient.", chanwideamt.ToFinString())
	}

	return nil
}

// 开始远程支付
func (s *Servicer) remotePay(newcur *Customer, msg *protocol.MsgRequestInitiatePayment, targetAddr *protocol.ChannelAccountAddress) error {
	return nil
}
