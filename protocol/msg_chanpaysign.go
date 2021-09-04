package protocol

import (
	"bytes"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
)

/**************************************************/

type MsgRequestChannelPayCollectionSign struct {
	OperationNum fields.VarUint8                      // 操作单号
	Bills        *channel.ChannelPayCompleteDocuments // 全部票据
}

func (m MsgRequestChannelPayCollectionSign) Type() uint8 {
	return MsgTypeRequestChannelPayCollectionSign
}

func (m MsgRequestChannelPayCollectionSign) Size() uint32 {
	return m.OperationNum.Size() + m.Bills.Size()
}

func (c MsgRequestChannelPayCollectionSign) Serialize() ([]byte, error) {
	var bt []byte
	var buffer bytes.Buffer
	bt, _ = c.OperationNum.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.Bills.Serialize() // 数据体
	buffer.Write(bt)
	return buffer.Bytes(), nil
}

func (c *MsgRequestChannelPayCollectionSign) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	// 通道
	seek, e = c.OperationNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = c.Bills.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	// 完成
	return seek, nil
}

func (m MsgRequestChannelPayCollectionSign) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

/**************************************************/

type MsgResponseChannelPayCollectionSign struct {
	OperationNum fields.VarUint8     // 操作单号
	ErrorMsg     fields.StringMax255 // 如果错误的消息
	Sign         fields.Sign         // 如果成功的签名
}

func (m MsgResponseChannelPayCollectionSign) Type() uint8 {
	return MsgTypeResponseChannelPayCollectionSign
}

func (m MsgResponseChannelPayCollectionSign) Size() uint32 {
	return m.OperationNum.Size() + m.ErrorMsg.Size() + m.Sign.Size()
}

func (c MsgResponseChannelPayCollectionSign) Serialize() ([]byte, error) {
	var bt []byte
	var buffer bytes.Buffer
	bt, _ = c.OperationNum.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.ErrorMsg.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.Sign.Serialize() // 数据体
	buffer.Write(bt)
	return buffer.Bytes(), nil
}

func (c *MsgResponseChannelPayCollectionSign) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = c.OperationNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = c.ErrorMsg.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = c.Sign.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	// 完成
	return seek, nil
}

func (m MsgResponseChannelPayCollectionSign) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

/**************************************************/

type MsgRequestChannelPayPaymentSign struct {
	OperationNum fields.VarUint8                      // 操作单号
	Bills        *channel.ChannelPayCompleteDocuments // 全部票据
}

func (m MsgRequestChannelPayPaymentSign) Type() uint8 {
	return MsgTypeRequestChannelPayPaymentSign
}

func (m MsgRequestChannelPayPaymentSign) Size() uint32 {
	return m.OperationNum.Size() + m.Bills.Size()
}

func (c MsgRequestChannelPayPaymentSign) Serialize() ([]byte, error) {
	var bt []byte
	var buffer bytes.Buffer
	bt, _ = c.OperationNum.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.Bills.Serialize() // 数据体
	buffer.Write(bt)
	return buffer.Bytes(), nil
}

func (c *MsgRequestChannelPayPaymentSign) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	// 通道
	seek, e = c.OperationNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = c.Bills.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	// 完成
	return seek, nil
}

func (m MsgRequestChannelPayPaymentSign) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

/**************************************************/

type MsgResponseChannelPayPaymentSign struct {
	OperationNum fields.VarUint8     // 操作单号
	ErrorMsg     fields.StringMax255 // 如果错误的消息
	Sign         fields.Sign         // 如果成功的签名
}

func (m MsgResponseChannelPayPaymentSign) Type() uint8 {
	return MsgTypeResponseChannelPayPaymentSign
}

func (m MsgResponseChannelPayPaymentSign) Size() uint32 {
	return m.OperationNum.Size() + m.ErrorMsg.Size() + m.Sign.Size()
}

func (c MsgResponseChannelPayPaymentSign) Serialize() ([]byte, error) {
	var bt []byte
	var buffer bytes.Buffer
	bt, _ = c.OperationNum.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.ErrorMsg.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.Sign.Serialize() // 数据体
	buffer.Write(bt)
	return buffer.Bytes(), nil
}

func (c *MsgResponseChannelPayPaymentSign) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = c.OperationNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = c.ErrorMsg.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = c.Sign.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	// 完成
	return seek, nil
}

func (m MsgResponseChannelPayPaymentSign) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

/**************************************************/

type MsgSendChannelPayCompletedSignaturesToDownstream struct {
	OperationNum fields.VarUint8       // 操作单号
	AllSigns     fields.SignListMax255 // 全部票据包含全部所需要的签名
}

func (m MsgSendChannelPayCompletedSignaturesToDownstream) Type() uint8 {
	return MsgTypeSendChannelPayCompletedSignedBillToDownstream
}

func (m MsgSendChannelPayCompletedSignaturesToDownstream) Size() uint32 {
	return m.OperationNum.Size() + m.AllSigns.Size()
}

func (c MsgSendChannelPayCompletedSignaturesToDownstream) Serialize() ([]byte, error) {
	var bt []byte
	var buffer bytes.Buffer
	bt, _ = c.OperationNum.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.AllSigns.Serialize() // 数据体
	buffer.Write(bt)
	return buffer.Bytes(), nil
}

func (c *MsgSendChannelPayCompletedSignaturesToDownstream) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	// 通道
	seek, e = c.OperationNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = c.AllSigns.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	// 完成
	return seek, nil
}

func (m MsgSendChannelPayCompletedSignaturesToDownstream) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
