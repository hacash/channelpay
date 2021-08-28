package protocol

import (
	"bytes"
)

/**
 * 通知客户端被顶替下线
 */

type MsgDisplacementOffline struct {
}

func (m MsgDisplacementOffline) Type() uint8 {
	return MsgTypeDisplacementOffline
}

func (m MsgDisplacementOffline) Size() uint32 {
	return 0
}

func (m *MsgDisplacementOffline) Parse(buf []byte, seek uint32) (uint32, error) {
	//var e error
	return seek, nil
}

func (m MsgDisplacementOffline) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	// ok
	return buf.Bytes(), nil
}

func (m MsgDisplacementOffline) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
