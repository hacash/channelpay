package chanpay

import (
	"bytes"
	"fmt"
	"github.com/hacash/chain/leveldb"
	"github.com/hacash/core/account"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"sync"
	"time"
)

/**
 * 本地 level db 数据库对所有数据储存器的实现
 */

type LocalDBImpOfDataSource struct {
	ldb *leveldb.DB

	accMapLock      sync.RWMutex
	tempPrivateKeys map[string]*account.Account
}

func NewLocalDBImpOfDataSource(dbdir string) (*LocalDBImpOfDataSource, error) {
	imp := &LocalDBImpOfDataSource{
		tempPrivateKeys: make(map[string]*account.Account),
	}

	db, e := leveldb.OpenFile(dbdir, nil)
	if e != nil {
		return nil, e
	}
	imp.ldb = db

	return imp, nil
}

// parse key
func (l *LocalDBImpOfDataSource) key(kind string, key []byte) []byte {
	buf := bytes.NewBuffer([]byte(kind))
	buf.Write(key)
	return buf.Bytes()
}

/********************************************************************/

func (l *LocalDBImpOfDataSource) Init() error {
	return nil
}

// 储存通用对账票据，检查对账单据的合法性
func (l *LocalDBImpOfDataSource) UpdateStoreBalanceBill(channelId fields.ChannelId, bill channel.ReconciliationBalanceBill) error {
	// data
	data, e := bill.SerializeWithTypeCode()
	if e != nil {
		return e
	}
	// save
	return l.ldb.Put(l.key("bill", channelId), data, nil)
}

// 读取最新票据
func (l *LocalDBImpOfDataSource) GetLastestBalanceBill(channelId fields.ChannelId) (channel.ReconciliationBalanceBill, error) {
	// save
	data, e := l.ldb.Get(l.key("bill", channelId), nil)
	if e != nil {
		return nil, nil // 不存在，未找到
	}
	// parse
	return channel.ParseReconciliationBalanceBillByPrefixTypeCode(data, 0)
}

/********************************************************************/

// 设定服务通道
func (l *LocalDBImpOfDataSource) setupChannel(channelId fields.ChannelId, weAreRightSide bool) error {
	key := l.key("chanset", channelId)
	vside := uint8(1)
	if weAreRightSide {
		vside = 2
	}
	return l.ldb.Put(key, []byte{vside}, nil)
}

// 查询服务通道
func (l *LocalDBImpOfDataSource) checkChannel(channelId fields.ChannelId) (bool, bool) {
	key := l.key("chanset", channelId)
	data, e := l.ldb.Get(key, nil)
	if e == nil && len(data) > 0 {
		weAreRightSide := false
		if data[0] == 2 {
			weAreRightSide = true
		}
		return true, weAreRightSide
	}
	// 未找到
	return false, false
}

// 取消服务通道
func (l *LocalDBImpOfDataSource) cancelChannel(channelId fields.ChannelId) error {
	key := l.key("chanset", channelId)
	return l.ldb.Delete(key, nil)
}

// 设定服务通道
func (l *LocalDBImpOfDataSource) SetupCustomerPayChannel(channelId fields.ChannelId) error {
	weAreRightSide := true
	return l.setupChannel(channelId, weAreRightSide)
}

// 查询服务通道
func (l *LocalDBImpOfDataSource) CheckCustomerPayChannel(channelId fields.ChannelId) bool {
	ok, _ := l.checkChannel(channelId)
	return ok
}

// 取消服务通道
func (l *LocalDBImpOfDataSource) CancelCustomerPayChannel(channelId fields.ChannelId) error {
	return l.cancelChannel(channelId)
}

/********************************************************************/

// 设定服务商结算通道，weAreRightSide 本方地址是否为右侧
func (l *LocalDBImpOfDataSource) SetupRelaySettlementPayChannel(channelId fields.ChannelId, weAreRightSide bool) error {
	return l.setupChannel(channelId, weAreRightSide)
}

// 查询服务商结算通道是否存在，前一个bool 表示是否存在，后一个bool=weIsRightSide
func (l *LocalDBImpOfDataSource) CheckRelaySettlementPayChannel(channelId fields.ChannelId) (bool, bool) {
	return l.checkChannel(channelId)
}

// 取消服务商结算通道
func (l *LocalDBImpOfDataSource) CancelRelaySettlementPayChannel(channelId fields.ChannelId) error {
	return l.cancelChannel(channelId)
}

/********************************************************************/
// 签名机

func (s *LocalDBImpOfDataSource) TemporaryStoragePrivateKeyForSign(privatekeyOrPassword string) {
	s.accMapLock.Lock()
	defer s.accMapLock.Unlock()
	acc := account.GetAccountByPrivateKeyOrPassword(privatekeyOrPassword)
	s.tempPrivateKeys[string(acc.Address)] = acc
}

// 移除私钥
func (s *LocalDBImpOfDataSource) RemovePrivateKey(address fields.Address) {
	s.accMapLock.Lock()
	defer s.accMapLock.Unlock()
	delete(s.tempPrivateKeys, string(address))
}

// 清除所有私钥
func (s *LocalDBImpOfDataSource) CleanAllPrivateKey() {
	s.accMapLock.Lock()
	defer s.accMapLock.Unlock()
	s.tempPrivateKeys = make(map[string]*account.Account)
}

func (s *LocalDBImpOfDataSource) CheckPaydocumentAndFillNeedSignature(paydocs *channel.ChannelPayCompleteDocuments, mustaddrs []fields.Address) (*fields.SignListMax255, error) {
	// 检查时间戳，不签署已经过期 60s 后的票据
	ctimes := time.Now().Unix()
	if int64(paydocs.ChainPayment.Timestamp) > ctimes {
		return nil, fmt.Errorf("The bill timestamp error")
	}
	if int64(paydocs.ChainPayment.Timestamp)+60 < ctimes {
		return nil, fmt.Errorf("The bill has expired and cannot be signed")
	}

	// 检查服务的通道
	// 找到两个通道，并且必须是顺序挨着的
	var paychan1 *channel.ChannelChainTransferProveBodyInfo = nil
	var paychan2 *channel.ChannelChainTransferProveBodyInfo = nil
	bodys := paydocs.ProveBodys.ProveBodys
	for i := 0; i < len(bodys)-1; i++ {
		// 必须两个连续的通道
		if hav1 := s.CheckCustomerPayChannel(bodys[i].ChannelId); hav1 {
			if hav2 := s.CheckCustomerPayChannel(bodys[i+1].ChannelId); hav2 {
				paychan1 = bodys[i]
				paychan2 = bodys[i+1]
			}
		}
	}
	if paychan1 == nil || paychan2 == nil {
		// 没找到支持的通道
		return nil, fmt.Errorf("Channel not support in check list.")
	}

	// 不做任何余额或者下游签名的检查，这些检查都放在外层
	// TODO:: 或者第三方签名机实现时再做必要的检查
	// 这里直接找出需要的私钥，直接签名

	// 填充签名，返回签名
	s.accMapLock.RLock()
	defer s.accMapLock.RUnlock()

	// 签名表
	var signs = fields.CreateEmptySignListMax255()

	// 取出私钥
	for _, v := range mustaddrs {
		if acc, hav := s.tempPrivateKeys[string(v)]; hav {
			sign, e := paydocs.ChainPayment.DoSignFillPosition(acc)
			if e != nil {
				return nil, e
			}
			signs.Append(*sign)
		} else {
			// 私钥没找到
			return nil, fmt.Errorf("Must sign address %s not find in sign machine.")
		}
	}

	// 签名成功，返回
	return signs, nil
}
