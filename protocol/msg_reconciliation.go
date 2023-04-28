package protocol

import (
	"bytes"
	"github.com/hacash/core/fields"
)

////////////////////////////////////////////

// Client initiated reconciliation
type MsgClientInitiateReconciliation struct {
	SelfSign fields.Sign // Reconciliation our signature
}

func (m MsgClientInitiateReconciliation) Type() uint8 {
	return MsgTypeClientInitiateReconciliation
}

func (m MsgClientInitiateReconciliation) Size() uint32 {
	return m.SelfSign.Size()
}

func (m *MsgClientInitiateReconciliation) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.SelfSign.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgClientInitiateReconciliation) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b1, e := m.SelfSign.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

func (m MsgClientInitiateReconciliation) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

////////////////////////////////////////////

// Server response reconciliation
type MsgServicerRespondReconciliation struct {
	SelfSign fields.Sign // Reconciliation our signature
}

func (m MsgServicerRespondReconciliation) Type() uint8 {
	return MsgTypeServicerRespondReconciliation
}

func (m MsgServicerRespondReconciliation) Size() uint32 {
	return m.SelfSign.Size()
}

func (m *MsgServicerRespondReconciliation) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.SelfSign.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgServicerRespondReconciliation) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b1, e := m.SelfSign.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

func (m MsgServicerRespondReconciliation) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
