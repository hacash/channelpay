package datasources

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
	UpdateStoreBalanceBill(channelId fields.Bytes16, bill channel.ReconciliationBalanceBill) error
	// 读取最新票据
	GetLastestBalanceBill(channelId fields.Bytes16) (channel.ReconciliationBalanceBill, error)
}

// 通道载入数据
type DataSourceOfServicerPayChannelSetup interface {
	Init() error // 初始化
	// 设定服务通道
	SetupServicerPayChannel(channelId fields.Bytes16) error
	// 查询服务通道
	CheckServicerPayChannel(channelId fields.Bytes16) bool
	// 取消服务通道
	CancelServicerPayChannel(channelId fields.Bytes16) error
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
