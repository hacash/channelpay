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
	NodeIdPath     *NodeIdPath
	PredictPathFee fields.Amount         // 路径预估手续费
	Describe       fields.StringMax65535 // 通道支付描述
}

func (m PayPathDescribe) Size() uint32 {
	return m.NodeIdPath.Size() + m.PredictPathFee.Size() + m.Describe.Size()
}

func (m *PayPathDescribe) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	m.NodeIdPath = &NodeIdPath{}
	seek, e = m.NodeIdPath.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.PredictPathFee.Parse(buf, seek)
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
	bt, e = m.PredictPathFee.Serialize()
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

/********************************************************/

/**
 * 预查询支付 ，请求
 */

type MsgRequestPrequeryPayment struct {
	PayAmount        fields.Amount       // 支付金额，必须为正整数
	PayeeChannelAddr fields.StringMax255 // 收款人通道地址，例如： 1Ke39SGbnrsDzkThANzTAFJmDhcc8qvM2Z__HACorg
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

/***************************************************/

// 支付预查询 响应
type MsgResponsePrequeryPayment struct {
	ErrCode   fields.VarUint2       // 当有错误时的错误码 > 0
	ErrTip    fields.StringMax65535 // 错误消息
	Notes     fields.StringMax65535 // 描述信息
	PathForms *PayPathForms         // 可选择的支付通道列表
}

func NewMsgResponsePrequeryPayment(ecode uint16) *MsgResponsePrequeryPayment {
	return &MsgResponsePrequeryPayment{
		ErrCode: fields.VarUint2(ecode),
	}
}

func (m MsgResponsePrequeryPayment) Type() uint8 {
	return MsgTypeResponsePrequeryPayment
}

func (m MsgResponsePrequeryPayment) Size() uint32 {
	size := m.ErrCode.Size()
	ecode := int(m.ErrCode)
	if ecode == 0 {
		size += m.Notes.Size() +
			m.PathForms.Size()
	} else {
		size += m.ErrTip.Size()
	}
	return size
}

func (m *MsgResponsePrequeryPayment) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.ErrCode.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	if int(m.ErrCode) == 0 {
		seek, e = m.Notes.Parse(buf, seek)
		if e != nil {
			return 0, e
		}
		m.PathForms = &PayPathForms{}
		seek, e = m.PathForms.Parse(buf, seek)
		if e != nil {
			return 0, e
		}
	} else {
		seek, e = m.ErrTip.Parse(buf, seek)
		if e != nil {
			return 0, e
		}
	}
	return seek, nil
}

func (m MsgResponsePrequeryPayment) Serialize() ([]byte, error) {
	var e error
	var bt []byte
	buf := bytes.NewBuffer(nil)
	bt, e = m.ErrCode.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	if int(m.ErrCode) == 0 {
		bt, e = m.Notes.Serialize()
		if e != nil {
			return nil, e
		}
		buf.Write(bt)
		bt, e = m.PathForms.Serialize()
		if e != nil {
			return nil, e
		}
		buf.Write(bt)
	} else {
		bt, e = m.ErrTip.Serialize()
		if e != nil {
			return nil, e
		}
		buf.Write(bt)
	}
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
