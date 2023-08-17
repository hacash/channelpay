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

type PayPathDescribe struct {
	NodeIdPath        *NodeIdPath
	PredictPathFeeAmt fields.Amount // Estimated service charge for route
	PredictPathFeeSat fields.SatoshiVariation
	Describe          fields.StringMax65535 // Channel payment description
}

func (m PayPathDescribe) Size() uint32 {
	return m.NodeIdPath.Size() + m.PredictPathFeeAmt.Size() + m.PredictPathFeeSat.Size() + m.Describe.Size()
}

func (m *PayPathDescribe) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	m.NodeIdPath = &NodeIdPath{}
	seek, e = m.NodeIdPath.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.PredictPathFeeAmt.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.PredictPathFeeSat.Parse(buf, seek)
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
	bt, e = m.PredictPathFeeAmt.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.PredictPathFeeSat.Serialize()
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

type PayPathForms struct {
	PayPathCount fields.VarUint1    // Number of payment paths
	PayPaths     []*PayPathDescribe // Payment path list
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
	PayAmount        fields.Amount           // Payment amount must be a positive integer
	PaySatoshi       fields.SatoshiVariation // bitcoin if
	PayeeChannelAddr fields.StringMax255     // Receiver channel address, for example: 1ke39sgbnrsdzkthanztafjmdhcc8qvm2z__ HACorg
}

func (m MsgRequestPrequeryPayment) Type() uint8 {
	return MsgTypeRequestPrequeryPayment
}

func (m MsgRequestPrequeryPayment) Size() uint32 {
	return m.PayAmount.Size() +
		m.PaySatoshi.Size() +
		m.PayeeChannelAddr.Size()
}

func (m *MsgRequestPrequeryPayment) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.PayAmount.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.PaySatoshi.Parse(buf, seek)
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
	b1, e := m.PayAmount.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	b2, e := m.PaySatoshi.Serialize()
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

// Payment pre query response
type MsgResponsePrequeryPayment struct {
	ErrCode   fields.VarUint2       // Error code when there is an error > 0
	ErrTip    fields.StringMax65535 // Error message
	Notes     fields.StringMax65535 // Descriptive information
	PathForms *PayPathForms         // List of selectable payment channels
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
