package protocol

import (
	"bytes"
)

/**
 * 通知客户端被顶替下线
 */

type MsgHeartbeat struct {
}

func (m MsgHeartbeat) Type() uint8 {
	return MsgTypeHeartbeat
}

func (m MsgHeartbeat) Size() uint32 {
	return 0
}

func (m *MsgHeartbeat) Parse(buf []byte, seek uint32) (uint32, error) {
	//var e error
	return seek, nil
}

func (m MsgHeartbeat) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	// ok
	return buf.Bytes(), nil
}

func (m MsgHeartbeat) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
