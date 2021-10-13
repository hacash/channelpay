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
	ProtocolVersion fields.VarUint2 // 服务端的最新协议版本号，用于提醒客户端更新软件版本
	BillIsExistent  fields.Bool     // 是否存在对账单
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
		// 必须带上 code
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
