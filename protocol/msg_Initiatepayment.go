package protocol

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**
 * 发起支付
 */
type MsgRequestInitiatePayment struct {
	TransactionDistinguishId fields.VarUint8 // Transaction identification ID

	Timestamp                fields.BlockTxTimestamp // Transaction timestamp
	OrderNoteHashHalfChecker fields.HashHalfChecker  // Order detail data hash

	HighestAcceptanceFee fields.Amount       // Maximum acceptable total handling fee amount
	PayAmount            fields.Amount       // Payment amount must be a positive integer
	PayeeChannelAddr     fields.StringMax255 // Receiver channel address, for example: 1ke39sgbnrsdzkthanztafjmdhcc8qvm2z__ HACorg

	// Specified routing node ID list
	TargetPath NodeIdPath
}

func (m MsgRequestInitiatePayment) Type() uint8 {
	return MsgTypeInitiatePayment
}

func (m MsgRequestInitiatePayment) Size() uint32 {
	return m.TransactionDistinguishId.Size() +
		m.Timestamp.Size() +
		m.OrderNoteHashHalfChecker.Size() +
		m.HighestAcceptanceFee.Size() +
		m.PayAmount.Size() +
		m.PayeeChannelAddr.Size() +
		m.TargetPath.Size()
}

func (m *MsgRequestInitiatePayment) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.TransactionDistinguishId.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.Timestamp.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
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

func (m MsgRequestInitiatePayment) Serialize() ([]byte, error) {
	var e error
	var bt []byte
	buf := bytes.NewBuffer(nil)
	bt, e = m.TransactionDistinguishId.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.Timestamp.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
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

func (m MsgRequestInitiatePayment) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

/**
 * 中继节点发起支付
 */
type MsgRequestRelayInitiatePayment struct {
	InitPayMsg         MsgRequestInitiatePayment
	IdentificationName fields.StringMax255
	ChannelId          fields.ChannelId
}

func (m MsgRequestRelayInitiatePayment) Type() uint8 {
	return MsgTypeRelayInitiatePayment
}

func (m MsgRequestRelayInitiatePayment) Serialize() ([]byte, error) {
	var e error
	var bt []byte
	buf := bytes.NewBuffer(nil)
	bt, e = m.InitPayMsg.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.IdentificationName.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.ChannelId.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	// ok
	return buf.Bytes(), nil
}

func (m MsgRequestRelayInitiatePayment) Size() uint32 {
	return m.InitPayMsg.Size() +
		m.IdentificationName.Size() +
		m.ChannelId.Size()
}

func (m *MsgRequestRelayInitiatePayment) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.InitPayMsg.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.IdentificationName.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.ChannelId.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgRequestRelayInitiatePayment) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
