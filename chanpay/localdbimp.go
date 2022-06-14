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

// Save general reconciliation bills and check the validity of reconciliation documents
func (l *LocalDBImpOfDataSource) UpdateStoreBalanceBill(channelId fields.ChannelId, bill channel.ReconciliationBalanceBill) error {
	// data
	data, e := bill.SerializeWithTypeCode()
	if e != nil {
		return e
	}
	// save
	return l.ldb.Put(l.key("bill", channelId), data, nil)
}

// Read latest ticket
func (l *LocalDBImpOfDataSource) GetLastestBalanceBill(channelId fields.ChannelId) (channel.ReconciliationBalanceBill, error) {
	// save
	data, e := l.ldb.Get(l.key("bill", channelId), nil)
	if e != nil {
		return nil, nil // Does not exist, not found
	}
	// parse
	bill, _, e := channel.ParseReconciliationBalanceBillByPrefixTypeCode(data, 0)
	return bill, e
}

/********************************************************************/

// Set service channel
func (l *LocalDBImpOfDataSource) setupChannel(channelId fields.ChannelId, weAreRightSide bool) error {
	key := l.key("chanset", channelId)
	vside := uint8(1)
	if weAreRightSide {
		vside = 2
	}
	return l.ldb.Put(key, []byte{vside}, nil)
}

// Query service channel
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
	// not found
	return false, false
}

// Cancel service channel
func (l *LocalDBImpOfDataSource) cancelChannel(channelId fields.ChannelId) error {
	key := l.key("chanset", channelId)
	return l.ldb.Delete(key, nil)
}

// Set service channel
func (l *LocalDBImpOfDataSource) SetupCustomerPayChannel(channelId fields.ChannelId) error {
	weAreRightSide := true
	return l.setupChannel(channelId, weAreRightSide)
}

// Query service channel
func (l *LocalDBImpOfDataSource) CheckCustomerPayChannel(channelId fields.ChannelId) bool {
	ok, _ := l.checkChannel(channelId)
	return ok
}

// Cancel service channel
func (l *LocalDBImpOfDataSource) CancelCustomerPayChannel(channelId fields.ChannelId) error {
	return l.cancelChannel(channelId)
}

/********************************************************************/

// Set the settlement channel of the service provider. Is our address on the right
func (l *LocalDBImpOfDataSource) SetupRelaySettlementPayChannel(channelId fields.ChannelId, weAreRightSide bool) error {
	return l.setupChannel(channelId, weAreRightSide)
}

// Query whether the settlement channel of the service provider exists. The previous bool indicates whether it exists, and the latter bool=weisrightside
func (l *LocalDBImpOfDataSource) CheckRelaySettlementPayChannel(channelId fields.ChannelId) (bool, bool) {
	return l.checkChannel(channelId)
}

// Cancel the settlement channel of the service provider
func (l *LocalDBImpOfDataSource) CancelRelaySettlementPayChannel(channelId fields.ChannelId) error {
	return l.cancelChannel(channelId)
}

/********************************************************************/
// Signature machine

func (s *LocalDBImpOfDataSource) TemporaryStoragePrivateKeyForSign(privatekeyOrPassword string) {
	s.accMapLock.Lock()
	defer s.accMapLock.Unlock()
	acc := account.GetAccountByPrivateKeyOrPassword(privatekeyOrPassword)
	s.tempPrivateKeys[string(acc.Address)] = acc
	//fmt.Println(acc.AddressReadable)
}

// Remove private key
func (s *LocalDBImpOfDataSource) RemovePrivateKey(address fields.Address) {
	s.accMapLock.Lock()
	defer s.accMapLock.Unlock()
	delete(s.tempPrivateKeys, string(address))
}

// Remove private key
func (s *LocalDBImpOfDataSource) ReadPrivateKey(address fields.Address) *account.Account {
	s.accMapLock.RLock()
	defer s.accMapLock.RUnlock()
	if acc, ok := s.tempPrivateKeys[string(address)]; ok {
		return acc
	}
	return nil
}

// Clear all private keys
func (s *LocalDBImpOfDataSource) CleanAllPrivateKey() {
	s.accMapLock.Lock()
	defer s.accMapLock.Unlock()
	s.tempPrivateKeys = make(map[string]*account.Account)
}

// Sign statement
func (s *LocalDBImpOfDataSource) CheckReconciliationFillNeedSignature(bill *channel.OffChainFormPaymentChannelRealtimeReconciliation, checksign *fields.Sign) (*fields.Sign, error) {
	var e error
	leftacc := s.ReadPrivateKey(bill.LeftAddress)
	var fillsign *fields.Sign = nil
	if leftacc != nil {
		// Fill in signature
		bill.RightSign = *checksign
		fillsign, _, e = bill.FillTargetSignature(leftacc)
		if e != nil {
			return nil, e
		}
	}
	if fillsign == nil {
		// Try the right side again
		rightacc := s.ReadPrivateKey(bill.RightAddress)
		if rightacc != nil {
			// Fill in signature
			bill.LeftSign = *checksign
			fillsign, _, e = bill.FillTargetSignature(rightacc)
			if e != nil {
				return nil, e
			}
		}
	}
	if fillsign == nil {
		return nil, fmt.Errorf("No address of private key that can be signed.")
	}
	// Check all signatures
	if e = bill.CheckAddressAndSign(); e != nil {
		return nil, fmt.Errorf("Signature verification failed.")
	}

	// success
	return fillsign, nil
}

// Signed payment
func (s *LocalDBImpOfDataSource) CheckPaydocumentAndFillNeedSignature(paydocs *channel.ChannelPayCompleteDocuments, mustaddrs []fields.Address) (*fields.SignListMax255, error) {
	// Check the time stamp, and do not sign bills that have expired for 60s
	ctimes := time.Now().Unix()
	if int64(paydocs.ChainPayment.Timestamp) > ctimes {
		return nil, fmt.Errorf("The bill timestamp error")
	}
	if int64(paydocs.ChainPayment.Timestamp)+60 < ctimes {
		return nil, fmt.Errorf("The bill has expired and cannot be signed")
	}

	// Check the channel of the service
	// Two channels are found and must be next to each other in sequence
	/*
		var paychan1 *channel.ChannelChainTransferProveBodyInfo = nil
		var paychan2 *channel.ChannelChainTransferProveBodyInfo = nil
		bodys := paydocs.ProveBodys.ProveBodys
		for i := 0; i < len(bodys)-1; i++ {
			// Must be two consecutive channels
			if hav1 := s.CheckCustomerPayChannel(bodys[i].ChannelId); hav1 {
				if hav2 := s.CheckCustomerPayChannel(bodys[i+1].ChannelId); hav2 {
					paychan1 = bodys[i]
					paychan2 = bodys[i+1]
				}
			}
		}
		if paychan1 == nil || paychan2 == nil {
			// No supported channels found
			return nil, fmt.Errorf("Channel not support in check list.")
		}
	*/

	// Do not check any balances or downstream signatures. These checks are placed on the outer layer
	// TODO:: 或者第三方签名机实现时再做必要的检查
	// Here, you can find the required private key and sign it directly

	// Fill in signature, return signature
	s.accMapLock.RLock()
	defer s.accMapLock.RUnlock()

	// Signature form
	var signs = fields.CreateEmptySignListMax255()

	// Remove the private key
	for _, v := range mustaddrs {
		if acc, hav := s.tempPrivateKeys[string(v)]; hav {
			sign, e := paydocs.ChainPayment.DoSignFillPosition(acc)
			if e != nil {
				return nil, e
			}
			signs.Append(*sign)
		} else {
			// Private key not found
			return nil, fmt.Errorf("Must sign address %s not find in sign machine account server list.", v.ToReadable())
		}
	}

	// Signature succeeded, return
	return signs, nil
}
