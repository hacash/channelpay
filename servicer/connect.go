package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
	"time"
)

// Client connection
func (s *Servicer) connectCustomerHandler(ws *websocket.Conn) {

	// Create customer connection
	customer := chanpay.NewCustomer(ws)

	// If not registered within 5 seconds, close the connection
	time.AfterFunc(time.Second*5, func() {
		if customer.RegisteredID == 0 {
			ws.Close() // Timeout unregistered, closed
		}
	})

	// Loop read message
	for {
		// Read message, parse message error, disconnect
		msgobj, _, err := protocol.ReceiveMsg(ws)
		if err != nil {
			break
		}
		// The first message must be a registration message
		if customer.RegisteredID == 0 {
			if msgobj.Type() == protocol.MsgTypeLogin {
				tarobj := msgobj.(*protocol.MsgLogin)
				customer.DoRegister(tarobj.ChannelId, tarobj.CustomerAddress) // Perform registration
				// Find old connections
				old := s.FindCustomersByChannel(tarobj.ChannelId)
				// Process old client connections
				if old != nil {
					// Remove from management pool
					s.RemoveCustomerFromPool(old)
					// Copy data, replace offline
					old.DoDisplacementOffline(customer)
				} else {
					// New login
					e := s.LoginNewCustomer(customer)
					if e != nil {
						// Login failed, send error message
						emsg := &protocol.MsgError{
							ErrCode: 1,
							ErrTip:  fields.CreateStringMax65535(e.Error()),
						}
						protocol.SendMsg(ws, emsg)
						// Disconnect
						break
					}
				}
				// Add new connection
				s.AddCustomerToPool(customer)
				// Send statement message
				billmsg := &protocol.MsgLoginCheckLastestBill{
					ProtocolVersion: fields.VarUint2(protocol.LatestProtocolVersion),
					BillIsExistent:  fields.CreateBool(false),
				}
				cusbill := customer.ChannelSide.GetReconciliationBill()
				if cusbill != nil {
					billmsg.BillIsExistent = fields.CreateBool(true)
					billmsg.LastBill = cusbill
				}
				protocol.SendMsg(customer.ChannelSide.WsConn, billmsg)
				// Continue to accept messages and subscribe to messages
				s.dealOtherMessage(customer)
				// Successful return
				break
			} else if msgobj.Type() == protocol.MsgTypeHeartbeat {
				// Heartbeat 
				customer.UpdateLastestHeartbeatTime() // Keep heartbeat alive
			} else {
				// Message type error, exit directly
				break
			}
		}

	}

	// Remove from management pool
	s.RemoveCustomerFromPool(customer)

	// Disconnect
	ws.Close()
}

// Relay payment service connection
func (s *Servicer) dealOtherMessage(customer *chanpay.Customer) {

	msgch := make(chan protocol.Message, 2)
	subobj := customer.ChannelSide.SubscribeMessage(msgch) // 订阅消息

	// Processing other types of messages
	for {
		select {
		case <-subobj.Err(): // 消息订阅错误
			// Remove from management pool
			s.RemoveCustomerFromPool(customer)
			// Disconnect
			customer.ChannelSide.WsConn.Close()
			return
		case msg := <-msgch:
			// Process other messages
			go s.msgHandler(customer, msg)
		}

	}

}

// Relay payment service connection
func (s *Servicer) connectRelayPayHandler(ws *websocket.Conn) {

	// Return error message
	errorReturn := func(e error) {
		errmsg := &protocol.MsgError{
			ErrCode: 0,
			ErrTip:  fields.CreateStringMax65535(e.Error()),
		}
		protocol.SendMsg(ws, errmsg)
		ws.Close() // to break off
	}

	isLaunchPay := false

	// If the initiate payment message is not received within 5 seconds, close the connection
	time.AfterFunc(time.Second*5, func() {
		if isLaunchPay == false {
			ws.Close() // Timeout unregistered, closed
		}
	})

	// Read message, parse message error, disconnect
	msgobj, _, err := protocol.ReceiveMsg(ws)
	if err != nil {
		ws.Close()
		return // to break off
	}

	mty := msgobj.Type()
	if mty == protocol.MsgTypeRelayInitiatePayment {
		// Initiate payment message
		initpaymsg, ok := msgobj.(*protocol.MsgRequestRelayInitiatePayment)
		if !ok {
			errorReturn(fmt.Errorf("MsgRequestRelayInitiatePayment format error"))
			return // Message parsing error
		}

		isLaunchPay = true                              // Mark has been set
		e := s.dealRelayInitiatePayment(ws, initpaymsg) // 处理支付
		if e != nil {
			errorReturn(fmt.Errorf("DealRelayInitiatePayment error: %s", e.Error()))
			return // error
		}

	} else {

		// Only accept remote payment messages
		errorReturn(fmt.Errorf("Msg type error"))
		return // error

	}

}
