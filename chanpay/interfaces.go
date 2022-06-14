package chanpay

import (
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
)

/**
 * 数据源接口
 */

// Balance bill data
type DataSourceOfBalanceBill interface {
	Init() error // initialization
	// Save general reconciliation bills and check the validity of reconciliation documents
	UpdateStoreBalanceBill(channelId fields.ChannelId, bill channel.ReconciliationBalanceBill) error
	// Read latest ticket
	GetLastestBalanceBill(channelId fields.ChannelId) (channel.ReconciliationBalanceBill, error)
}

// Channel load data
type DataSourceOfServicerPayChannelSetup interface {
	Init() error // initialization
	// Set customer service channel
	SetupCustomerPayChannel(channelId fields.ChannelId) error
	// Query whether the customer service channel exists,
	CheckCustomerPayChannel(channelId fields.ChannelId) bool
	// Cancel customer service channel
	CancelCustomerPayChannel(channelId fields.ChannelId) error

	// Set the settlement channel of the service provider, and whether the weisrightside address is on the right
	//SetupRelaySettlementPayChannel(channelId fields.ChannelId, weIsRightSide bool) error
	// Query whether the settlement channel of the service provider exists. The previous bool indicates whether it exists, and the latter bool=weisrightside
	//CheckRelaySettlementPayChannel(channelId fields.ChannelId) (bool, bool)
	// Cancel the settlement channel of the service provider
	//CancelRelaySettlementPayChannel(channelId fields.ChannelId) error
}

// Signature machine
type DataSourceOfSignatureMachine interface {
	Init() error // initialization
	// Temporary private key
	TemporaryStoragePrivateKeyForSign(privatekeyOrPassword string)
	RemovePrivateKey(address fields.Address) // Remove private key
	CleanAllPrivateKey()                     // Clear all private keys
	// Sign the statement and then check all signatures
	CheckReconciliationFillNeedSignature(bill *channel.OffChainFormPaymentChannelRealtimeReconciliation, checksign *fields.Sign) (*fields.Sign, error)
	// Send the channel transaction to the signer to verify the data, and automatically fill in the signature
	CheckPaydocumentAndFillNeedSignature(paydocs *channel.ChannelPayCompleteDocuments, mustaddrs []fields.Address) (*fields.SignListMax255, error)
}
