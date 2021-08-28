package datasources

import (
	"bytes"
	"github.com/hacash/chain/leveldb"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
)

/**
 * 本地 level db 数据库对所有数据储存器的实现
 */

type LocalDBImpOfDataSource struct {
	ldb *leveldb.DB
}

func NewLocalDBImpOfDataSource(dbdir string) (*LocalDBImpOfDataSource, error) {
	imp := &LocalDBImpOfDataSource{}

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
func (l *LocalDBImpOfDataSource) UpdateStoreBalanceBill(channelId fields.Bytes16, bill channel.ReconciliationBalanceBill) error {
	// data
	data, e := bill.SerializeWithTypeCode()
	if e != nil {
		return e
	}
	// save
	return l.ldb.Put(l.key("bill", channelId), data, nil)
}

// 读取最新票据
func (l *LocalDBImpOfDataSource) GetLastestBalanceBill(channelId fields.Bytes16) (channel.ReconciliationBalanceBill, error) {
	// save
	data, e := l.ldb.Get(l.key("bill", channelId), nil)
	if e != nil {
		return nil, e
	}
	// parse
	return channel.ParseReconciliationBalanceBillByPrefixTypeCode(data, 0)
}

/********************************************************************/

// 设定服务通道
func (l *LocalDBImpOfDataSource) SetupServicerPayChannel(channelId fields.Bytes16) error {
	key := l.key("chanset", channelId)
	return l.ldb.Put(key, []byte{1}, nil)
}

// 查询服务通道
func (l *LocalDBImpOfDataSource) CheckServicerPayChannel(channelId fields.Bytes16) bool {
	key := l.key("chanset", channelId)
	data, e := l.ldb.Get(key, nil)
	if e == nil && len(data) > 0 {
		return true
	}
	// 未找到
	return false
}

// 取消服务通道
func (l *LocalDBImpOfDataSource) CancelServicerPayChannel(channelId fields.Bytes16) error {
	key := l.key("chanset", channelId)
	return l.ldb.Delete(key, nil)
}
