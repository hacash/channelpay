package protocol

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**
 * 客户端退出登录
 */

type MsgCustomerLogout struct {
	PostBack fields.StringMax255 // Return message
}

func (m MsgCustomerLogout) Type() uint8 {
	return MsgTypeLogout
}

func (m MsgCustomerLogout) Size() uint32 {
	return m.PostBack.Size()
}

func (m *MsgCustomerLogout) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.PostBack.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgCustomerLogout) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b1, e := m.PostBack.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

func (m MsgCustomerLogout) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
