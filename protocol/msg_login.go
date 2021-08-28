package protocol

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**
客户端连接注册消息
*/

type MsgLogin struct {
	ProtocolVersion fields.VarUint2     // 客户端协议版本号，用于升级和向前兼容
	ChannelId       fields.Bytes16      // 通道id
	CustomerAddress fields.Address      // 客户侧地址
	LanguageSet     fields.StringMax255 // 语言设置
}

func (m MsgLogin) Type() uint8 {
	return MsgTypeLogin
}

func (m MsgLogin) Size() uint32 {
	return m.ProtocolVersion.Size() +
		m.ChannelId.Size() +
		m.CustomerAddress.Size() +
		m.LanguageSet.Size()

}

func (m *MsgLogin) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.ProtocolVersion.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.ChannelId.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.CustomerAddress.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.LanguageSet.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgLogin) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b1, e := m.ProtocolVersion.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	b2, e := m.ChannelId.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b2)
	b3, e := m.CustomerAddress.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b3)
	b4, e := m.LanguageSet.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b4)
	// ok
	return buf.Bytes(), nil
}

func (m MsgLogin) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
