package protocol

import (
	"bytes"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
)

// 发起远程支付消息

type MsgRequestLaunchRemoteChannelPayment struct {
}

func (m MsgRequestLaunchRemoteChannelPayment) Type() uint8 {
	return MsgTypeRequestLaunchRemoteChannelPayment
}

//////////////////////////////////////////////////////

// 终端节点响应支付消息
type MsgResponseRemoteChannelPayment struct {
	OperationNum fields.VarUint8 // 操作单号
	// 包含错误
	ErrorCode fields.VarUint2
	ErrorMsg  fields.StringMax255
	// 正常响应
	ChannelPayBodyList *channel.ChannelPayProveBodyList
}

func (m MsgResponseRemoteChannelPayment) Type() uint8 {
	return MsgTypeResponseRemoteChannelPayment
}

///////////////////////////////////////////////////////

type MsgRequestRemoteChannelPayCollectionSign struct {
	OperationNum fields.VarUint8                      // 操作单号
	Bills        *channel.ChannelPayCompleteDocuments // 全部票据
}

func (m MsgRequestRemoteChannelPayCollectionSign) Type() uint8 {
	return MsgTypeRequestRemoteChannelPayCollectionSign
}

func (m MsgRequestRemoteChannelPayCollectionSign) Size() uint32 {
	return m.OperationNum.Size() + m.Bills.Size()
}

func (c MsgRequestRemoteChannelPayCollectionSign) Serialize() ([]byte, error) {
	var bt []byte
	var buffer bytes.Buffer
	bt, _ = c.OperationNum.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.Bills.Serialize() // 数据体
	buffer.Write(bt)
	return buffer.Bytes(), nil
}

func (c *MsgRequestRemoteChannelPayCollectionSign) Parse(buf []byte, seek uint32) (uint32, error) {
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

func (m MsgRequestRemoteChannelPayCollectionSign) SerializeWithType() ([]byte, error) {
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

type MsgResponseRemoteChannelPayCollectionSign struct {
	OperationNum fields.VarUint8     // 操作单号
	ErrorCode    fields.VarUint2     // 如果错误的消息
	ErrorMsg     fields.StringMax255 // 如果错误的消息
	Sign         fields.Sign         // 如果成功的签名
}

func (m MsgResponseRemoteChannelPayCollectionSign) Type() uint8 {
	return MsgTypeResponseRemoteChannelPayCollectionSign
}

func (m MsgResponseRemoteChannelPayCollectionSign) Size() uint32 {
	return m.OperationNum.Size() + m.ErrorCode.Size() + m.ErrorMsg.Size() + m.Sign.Size()
}

func (c MsgResponseRemoteChannelPayCollectionSign) Serialize() ([]byte, error) {
	var bt []byte
	var buffer bytes.Buffer
	bt, _ = c.OperationNum.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.ErrorCode.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.ErrorMsg.Serialize() // 数据体
	buffer.Write(bt)
	bt, _ = c.Sign.Serialize() // 数据体
	buffer.Write(bt)
	return buffer.Bytes(), nil
}

func (c *MsgResponseRemoteChannelPayCollectionSign) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = c.OperationNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = c.ErrorCode.Parse(buf, seek)
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

func (m MsgResponseRemoteChannelPayCollectionSign) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
