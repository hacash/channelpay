package protocol

import (
	"bytes"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
)

/**
 * 检查最新对账单
 */

type MsgLoginCheckLastestBill struct {
	ProtocolVersion fields.VarUint2 // The latest protocol version number of the server, which is used to remind the client to update the software version
	BillIsExistent  fields.Bool     // Whether there is a statement
	LastBill        channel.ReconciliationBalanceBill
}

func (m MsgLoginCheckLastestBill) Type() uint8 {
	return MsgTypeLoginCheckLastestBill
}

func (m MsgLoginCheckLastestBill) Size() uint32 {
	size := m.ProtocolVersion.Size() +
		m.BillIsExistent.Size()
	if m.BillIsExistent.Check() {
		size += m.LastBill.Size()
	}
	return size
}

func (m *MsgLoginCheckLastestBill) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.ProtocolVersion.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.BillIsExistent.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	if m.BillIsExistent.Check() {
		m.LastBill, seek, e = channel.ParseReconciliationBalanceBillByPrefixTypeCode(buf, seek)
		if e != nil {
			return 0, e
		}
	}
	return seek, nil
}

func (m MsgLoginCheckLastestBill) Serialize() ([]byte, error) {
	var e error
	var bt []byte = nil
	buf := bytes.NewBuffer(nil)
	bt, e = m.ProtocolVersion.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.BillIsExistent.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	if m.BillIsExistent.Check() {
		// Code is required
		bt, e = m.LastBill.SerializeWithTypeCode()
		if e != nil {
			return nil, e
		}
		buf.Write(bt)
	}
	// ok
	return buf.Bytes(), nil
}

func (m MsgLoginCheckLastestBill) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
