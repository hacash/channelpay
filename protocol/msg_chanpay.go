package protocol

import (
	"bytes"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
)

/**
 * 支付相关
 */

// 广播对账单
type MsgBroadcastChannelStatementProveBody struct {
	TransactionDistinguishId fields.VarUint8 // 交易识别id
	ProveBodyIndex           fields.VarUint1 // 对账单位置 index 从零开始
	ProveBodyInfo            *channel.ChannelChainTransferProveBodyInfo
}

func (m MsgBroadcastChannelStatementProveBody) Type() uint8 {
	return MsgTypeBroadcastChannelStatementProveBody
}

func (m MsgBroadcastChannelStatementProveBody) Size() uint32 {
	return m.TransactionDistinguishId.Size() + m.ProveBodyIndex.Size() + m.ProveBodyInfo.Size()
}

func (m *MsgBroadcastChannelStatementProveBody) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.TransactionDistinguishId.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.ProveBodyIndex.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	m.ProveBodyInfo = new(channel.ChannelChainTransferProveBodyInfo)
	seek, e = m.ProveBodyInfo.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgBroadcastChannelStatementProveBody) Serialize() ([]byte, error) {
	var e error
	var bt []byte = nil
	buf := bytes.NewBuffer(nil)
	bt, e = m.TransactionDistinguishId.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.ProveBodyIndex.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.ProveBodyInfo.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	// ok
	return buf.Bytes(), nil
}

func (m MsgBroadcastChannelStatementProveBody) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

/////////////////////////////////////////////////////////////

// 广播通道支付签名
type MsgBroadcastChannelStatementSignature struct {
	TransactionDistinguishId fields.VarUint8 // 交易识别id
	// 签名
	Signs fields.SignListMax255
}

func (MsgBroadcastChannelStatementSignature) Type() uint8 {
	return MsgTypeBroadcastChannelStatementSignature
}

func (m MsgBroadcastChannelStatementSignature) Size() uint32 {
	return m.TransactionDistinguishId.Size() + m.Signs.Size()
}

func (m *MsgBroadcastChannelStatementSignature) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.TransactionDistinguishId.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.Signs.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgBroadcastChannelStatementSignature) Serialize() ([]byte, error) {
	var e error
	var bt []byte = nil
	buf := bytes.NewBuffer(nil)
	bt, e = m.TransactionDistinguishId.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.Signs.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	// ok
	return buf.Bytes(), nil
}

func (m MsgBroadcastChannelStatementSignature) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

//////////////////////////////////////////////////////////////

// 通道支付错误
type MsgBroadcastChannelStatementError struct {
	ErrCode fields.VarUint2
	ErrTip  fields.StringMax65535
}

func (m MsgBroadcastChannelStatementError) Type() uint8 {
	return MsgTypeBroadcastChannelStatementError
}

func (m MsgBroadcastChannelStatementError) Size() uint32 {
	return m.ErrCode.Size() + m.ErrTip.Size()
}

func (m *MsgBroadcastChannelStatementError) Parse(buf []byte, seek uint32) (uint32, error) {
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

func (m MsgBroadcastChannelStatementError) Serialize() ([]byte, error) {
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

func (m MsgBroadcastChannelStatementError) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

//////////////////////////////////////////////////////////////

// 通道支付成功
type MsgBroadcastChannelStatementSuccessed struct {
	SuccessTip fields.StringMax65535
}

func (m MsgBroadcastChannelStatementSuccessed) Type() uint8 {
	return MsgTypeBroadcastChannelStatementSuccessed
}

func (m MsgBroadcastChannelStatementSuccessed) Size() uint32 {
	return m.SuccessTip.Size()
}

func (m *MsgBroadcastChannelStatementSuccessed) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.SuccessTip.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgBroadcastChannelStatementSuccessed) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b2, e := m.SuccessTip.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b2)
	// ok
	return buf.Bytes(), nil
}

func (m MsgBroadcastChannelStatementSuccessed) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
