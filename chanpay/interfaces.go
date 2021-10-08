package chanpay

import (
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
)

/**
 * 数据源接口
 */

// 余额票据数据
type DataSourceOfBalanceBill interface {
	Init() error // 初始化
	// 储存通用对账票据，检查对账单据的合法性
	UpdateStoreBalanceBill(channelId fields.ChannelId, bill channel.ReconciliationBalanceBill) error
	// 读取最新票据
	GetLastestBalanceBill(channelId fields.ChannelId) (channel.ReconciliationBalanceBill, error)
}

// 通道载入数据
type DataSourceOfServicerPayChannelSetup interface {
	Init() error // 初始化
	// 设定客户服务通道
	SetupCustomerPayChannel(channelId fields.ChannelId) error
	// 查询客户服务通道是否存在，
	CheckCustomerPayChannel(channelId fields.ChannelId) bool
	// 取消客户服务通道
	CancelCustomerPayChannel(channelId fields.ChannelId) error

	// 设定服务商结算通道，weIsRightSide 本方地址是否为右侧
	//SetupRelaySettlementPayChannel(channelId fields.ChannelId, weIsRightSide bool) error
	// 查询服务商结算通道是否存在，前一个bool 表示是否存在，后一个bool=weIsRightSide
	//CheckRelaySettlementPayChannel(channelId fields.ChannelId) (bool, bool)
	// 取消服务商结算通道
	//CancelRelaySettlementPayChannel(channelId fields.ChannelId) error
}

// 签名机
type DataSourceOfSignatureMachine interface {
	Init() error // 初始化
	// 暂存私钥
	TemporaryStoragePrivateKeyForSign(privatekeyOrPassword string)
	RemovePrivateKey(address fields.Address) // 移除私钥
	CleanAllPrivateKey()                     // 清除所有私钥
	// 将通道交易送入签名机验证数据，并自动填充签名
	CheckPaydocumentAndFillNeedSignature(paydocs *channel.ChannelPayCompleteDocuments, mustaddrs []fields.Address) (*fields.SignListMax255, error)
}
