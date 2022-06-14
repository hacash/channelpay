package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
)

// Initialize channel
func (s *Servicer) InitializeChannelSide(side *chanpay.ChannelSideConn, remoteAddress fields.Address, weIsLeft bool) error {
	if side == nil || side.ChannelId == nil {
		return fmt.Errorf("side or side.ChannelId is nil")
	}
	cid := side.ChannelId

	// Read channel status
	chanInfo, e := protocol.RequestRpcReqChannelInfo(s.config.FullNodeRpcUrl, cid)
	if e != nil {
		return fmt.Errorf("request channel info fail: %s", e.Error())
	}

	// Check channel status
	if chanInfo.Status != stores.ChannelStatusOpening {
		return fmt.Errorf("channel status is not on opening!")
	}

	var remoteIsRight = true

	// Query whether it is in the channel service list
	if remoteAddress != nil {
		// Customer service channel
		serhav := s.chanset.CheckCustomerPayChannel(cid)
		if serhav == false {
			// Not in service list
			return fmt.Errorf("channel %s is not in the service list.", cid.ToHex())
		}
		remoteIsLeft := remoteAddress.Equal(chanInfo.LeftAddress)
		remoteIsRight = remoteAddress.Equal(chanInfo.RightAddress)
		if !remoteIsLeft && !remoteIsRight {
			return fmt.Errorf("remoteAddress %s is not in the channel addresses.", remoteAddress.ToReadable())
		}
	} else {
		// Relay settlement node channel
		if weIsLeft {
			remoteAddress = chanInfo.RightAddress
			remoteIsRight = true
		} else {
			remoteAddress = chanInfo.LeftAddress
			remoteIsRight = false
		}
	}

	// handle
	// Read the latest statement
	bill, e := s.billstore.GetLastestBalanceBill(cid)
	if e != nil {
		return fmt.Errorf("load lastest balance bill error: %s", e.Error())
	}
	var chanLeftAddr = chanInfo.LeftAddress
	var chanRightAddr = chanInfo.RightAddress
	if bill != nil {
		side.SetReconciliationBill(bill)
		if chanLeftAddr.NotEqual(bill.GetLeftAddress()) ||
			chanRightAddr.NotEqual(bill.GetRightAddress()) {
			// Ticket and channel address do not match
			return fmt.Errorf("channel and bill address not match.")
		}
	}

	// Channel and account information
	side.ChannelInfo = chanInfo // Channel info
	if remoteIsRight {
		side.OurAddress = chanLeftAddr
		side.RemoteAddress = chanRightAddr
	} else {
		side.OurAddress = chanRightAddr
		side.RemoteAddress = chanLeftAddr
	}

	return nil
}

// New login to a client
func (s *Servicer) LoginNewCustomer(newcur *chanpay.Customer) error {
	// initialization
	e := s.InitializeChannelSide(newcur.ChannelSide, newcur.ChannelSide.RemoteAddress, false)
	if e != nil {
		return e
	}

	// Start listening for messages
	newcur.ChannelSide.StartMessageListen()

	// Login successful
	return nil
}
