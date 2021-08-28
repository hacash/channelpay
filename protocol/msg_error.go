package protocol

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**
 * 错误消息
 */

type MsgError struct {
	ErrCode fields.VarUint2
	ErrTip  fields.StringMax65535
}

func (m MsgError) Type() uint8 {
	return MsgTypeError
}

func (m MsgError) Size() uint32 {
	return m.ErrCode.Size() + m.ErrTip.Size()
}

func (m *MsgError) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.ErrCode.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.ErrTip.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgError) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b1, e := m.ErrCode.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	b2, e := m.ErrTip.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b2)
	// ok
	return buf.Bytes(), nil
}

func (m MsgError) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
