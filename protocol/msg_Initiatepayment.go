package protocol

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**
 * 发起支付
 */

type MsgInitiatePayment struct {
	OrderNoteHashHalfChecker fields.HashHalfChecker // 订单详情数据哈希

	HighestAcceptanceFee fields.Amount       // 最高可接受的总手续费数额
	PayAmount            fields.Amount       // 支付金额，必须为正整数
	PayeeChannelAddr     fields.StringMax255 // 收款人通道地址，例如： 1Ke39SGbnrsDzkThANzTAFJmDhcc8qvM2Z__HACorg

	// 指定的路由节点ID列表
	TargetPath NodeIdPath
}

func (m MsgInitiatePayment) Type() uint8 {
	return MsgTypeInitiatePayment
}

func (m MsgInitiatePayment) Size() uint32 {
	return m.OrderNoteHashHalfChecker.Size() +
		m.HighestAcceptanceFee.Size() +
		m.PayAmount.Size() +
		m.PayeeChannelAddr.Size() +
		m.TargetPath.Size()
}

func (m *MsgInitiatePayment) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.OrderNoteHashHalfChecker.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.HighestAcceptanceFee.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.PayAmount.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.PayeeChannelAddr.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.TargetPath.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgInitiatePayment) Serialize() ([]byte, error) {
	var e error
	var bt []byte
	buf := bytes.NewBuffer(nil)
	bt, e = m.OrderNoteHashHalfChecker.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.HighestAcceptanceFee.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.PayAmount.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.PayeeChannelAddr.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.TargetPath.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	// ok
	return buf.Bytes(), nil
}

func (m MsgInitiatePayment) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
