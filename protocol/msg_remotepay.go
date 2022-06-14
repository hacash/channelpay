package protocol

// Initiate remote payment message
/*
type MsgRequestLaunchRemoteChannelPayment struct {
	OrderNoteHashHalfChecker fields.HashHalfChecker // Order detail data hash

	HighestAcceptanceFee fields.Amount       // Maximum acceptable total handling fee amount
	PayAmount            fields.Amount       // Payment amount must be a positive integer
	PayeeChannelAddr     fields.StringMax255 // Receiver channel address, for example: 1ke39sgbnrsdzkthanztafjmdhcc8qvm2z__ HACorg

	// Specified routing node ID list
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

// Terminal node responds to payment message
type MsgResponseRemoteChannelPayment struct {
	OperationNum fields.VarUint8 // Operation No
	// Contains errors
	ErrorCode fields.VarUint2
	ErrorMsg  fields.StringMax255
	// Normal response
	ChannelPayBodyList *channel.ChannelPayProveBodyList
}

func (m MsgResponseRemoteChannelPayment) Type() uint8 {
	return MsgTypeResponseRemoteChannelPayment
}

///////////////////////////////////////////////////////

type MsgRequestRemoteChannelPayCollectionSign struct {
	OperationNum fields.VarUint8                      // Operation No
	Bills        *channel.ChannelPayCompleteDocuments // All bills
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
	bt, _ = c.OperationNum.Serialize() // Data body
	buffer.Write(bt)
	bt, _ = c.Bills.Serialize() // Data body
	buffer.Write(bt)
	return buffer.Bytes(), nil
}

func (c *MsgRequestRemoteChannelPayCollectionSign) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	// passageway
	seek, e = c.OperationNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = c.Bills.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	// complete
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
	OperationNum fields.VarUint8     // Operation No
	ErrorCode    fields.VarUint2     // If the wrong message
	ErrorMsg     fields.StringMax255 // If the wrong message
	Sign         fields.Sign         // If successful signature
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
	bt, _ = c.OperationNum.Serialize() // Data body
	buffer.Write(bt)
	bt, _ = c.ErrorCode.Serialize() // Data body
	buffer.Write(bt)
	bt, _ = c.ErrorMsg.Serialize() // Data body
	buffer.Write(bt)
	bt, _ = c.Sign.Serialize() // Data body
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
	// complete
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
