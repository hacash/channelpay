package protocol

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**
 * 预查询支付，请求
 */

type NodeIdPath struct {
	NodeIdCount fields.VarUint1
	NodeIdPath  []fields.VarUint4
}

func (m NodeIdPath) Size() uint32 {
	return m.NodeIdCount.Size() +
		uint32(m.NodeIdCount)*4
}

func (m *NodeIdPath) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.NodeIdCount.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	m.NodeIdPath = make([]fields.VarUint4, int(m.NodeIdCount))
	for i := 0; i < int(m.NodeIdCount); i++ {
		m.NodeIdPath[i] = fields.VarUint4(0)
		seek, e = m.NodeIdPath[i].Parse(buf, seek)
		if e != nil {
			return 0, e
		}
	}
	return seek, nil
}

func (m NodeIdPath) Serialize() ([]byte, error) {
	var e error = nil
	var bt []byte = nil
	buf := bytes.NewBuffer(nil)
	bt, e = m.NodeIdCount.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	for i := 0; i < int(m.NodeIdCount); i++ {
		bt, e = m.NodeIdPath[i].Serialize()
		if e != nil {
			return nil, e
		}
		buf.Write(bt)
	}
	// ok
	return buf.Bytes(), nil
}

/***************************************************/

//
type PayPathDescribe struct {
	NodeIdPath *NodeIdPath
	Describe   fields.StringMax65535 // 通道支付描述
}

func (m PayPathDescribe) Size() uint32 {
	return m.NodeIdPath.Size() + m.Describe.Size()
}

func (m *PayPathDescribe) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.NodeIdPath.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.Describe.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m PayPathDescribe) Serialize() ([]byte, error) {
	var e error = nil
	var bt []byte = nil
	buf := bytes.NewBuffer(nil)
	bt, e = m.NodeIdPath.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.Describe.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	// ok
	return buf.Bytes(), nil
}

/***************************************************/

//
type PayPathForms struct {
	PayPathCount fields.VarUint1    // 支付路径数
	PayPaths     []*PayPathDescribe // 支付路径列表
}

func (m PayPathForms) Size() uint32 {
	size := m.PayPathCount.Size()
	for i := 0; i < int(m.PayPathCount); i++ {
		size += m.PayPaths[i].Size()
	}
	return size
}

func (m *PayPathForms) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.PayPathCount.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	m.PayPaths = make([]*PayPathDescribe, int(m.PayPathCount))
	for i := 0; i < int(m.PayPathCount); i++ {
		m.PayPaths[i] = new(PayPathDescribe)
		seek, e = m.PayPaths[i].Parse(buf, seek)
		if e != nil {
			return 0, e
		}
	}
	return seek, nil
}

func (m PayPathForms) Serialize() ([]byte, error) {
	var e error = nil
	var bt []byte = nil
	buf := bytes.NewBuffer(nil)
	bt, e = m.PayPathCount.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	for i := 0; i < int(m.PayPathCount); i++ {
		bt, e = m.PayPaths[i].Serialize()
		if e != nil {
			return nil, e
		}
		buf.Write(bt)
	}
	// ok
	return buf.Bytes(), nil
}

/***************************************************/

// 支付预查询
type MsgRequestPrequeryPayment struct {
	PredictTotalFee  fields.Amount       // 总共需要支付的手续费
	PayAmount        fields.Amount       // 支付金额，必须为正整数
	PayeeChannelAddr fields.StringMax255 // 收款人通道地址，例如： 1Ke39SGbnrsDzkThANzTAFJmDhcc8qvM2Z__HACorg

	Notes     fields.StringMax255 // 描述信息
	PathForms *PayPathForms
}

func (m MsgRequestPrequeryPayment) Type() uint8 {
	return MsgTypeRequestPrequeryPayment
}

func (m MsgRequestPrequeryPayment) Size() uint32 {
	return m.PayAmount.Size() +
		m.PayeeChannelAddr.Size()
}

func (m *MsgRequestPrequeryPayment) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.PayAmount.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.PayeeChannelAddr.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgRequestPrequeryPayment) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b2, e := m.PayAmount.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b2)
	b3, e := m.PayeeChannelAddr.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b3)
	// ok
	return buf.Bytes(), nil
}

func (m MsgRequestPrequeryPayment) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

/********************************************************/

/**
 * 预查询支付 ，内容响应
 */

type MsgResponsePrequeryPayment struct {
	PayAmount        fields.Amount       // 支付金额，必须为正整数
	PayeeChannelAddr fields.StringMax255 // 收款人通道地址，例如： 1Ke39SGbnrsDzkThANzTAFJmDhcc8qvM2Z__HACorg
}

func (m MsgResponsePrequeryPayment) Type() uint8 {
	return MsgTypeResponsePrequeryPayment
}

func (m MsgResponsePrequeryPayment) Size() uint32 {
	return m.PayAmount.Size() +
		m.PayeeChannelAddr.Size()
}

func (m *MsgResponsePrequeryPayment) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.PayAmount.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.PayeeChannelAddr.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgResponsePrequeryPayment) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b2, e := m.PayAmount.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b2)
	b3, e := m.PayeeChannelAddr.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b3)
	// ok
	return buf.Bytes(), nil
}

func (m MsgResponsePrequeryPayment) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
