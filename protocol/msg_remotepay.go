package protocol

// 发起远程支付消息
/*
type MsgRequestLaunchRemoteChannelPayment struct {
	OrderNoteHashHalfChecker fields.HashHalfChecker // 订单详情数据哈希

	HighestAcceptanceFee fields.Amount       // 最高可接受的总手续费数额
	PayAmount            fields.Amount       // 支付金额，必须为正整数
	PayeeChannelAddr     fields.StringMax255 // 收款人通道地址，例如： 1Ke39SGbnrsDzkThANzTAFJmDhcc8qvM2Z__HACorg

	// 指定的路由节点ID列表
	TargetPath NodeIdPath
}

func (m MsgRequestLaunchRemoteChannelPayment) Type() uint8 {
	return MsgTypeRequestLaunchRemoteChannelPayment
}

func (m MsgRequestLaunchRemoteChannelPayment) Size() uint32 {
	return m.OrderNoteHashHalfChecker.Size() +
		m.HighestAcceptanceFee.Size() +
		m.PayAmount.Size() +
		m.PayeeChannelAddr.Size() +
		m.TargetPath.Size()
}

func (m *MsgRequestLaunchRemoteChannelPayment) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
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

func (m MsgRequestLaunchRemoteChannelPayment) Serialize() ([]byte, error) {
	var e error
	var bt []byte
	buf := bytes.NewBuffer(nil)
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

func (m MsgRequestLaunchRemoteChannelPayment) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

func (m *MsgRequestLaunchRemoteChannelPayment) CopyFromInitiatePayment(msg *MsgRequestInitiatePayment) {
	m.OrderNoteHashHalfChecker = msg.OrderNoteHashHalfChecker
	m.HighestAcceptanceFee = msg.HighestAcceptanceFee
	m.PayAmount = msg.PayAmount
	m.PayeeChannelAddr = msg.PayeeChannelAddr
	m.TargetPath = msg.TargetPath
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

/**************************************************

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

*/
