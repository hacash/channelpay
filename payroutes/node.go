package payroutes

import (
	"bytes"
	"github.com/hacash/core/fields"
	"math/big"
)

/**
 * 支付路由节点
 */
type PayRelayNode struct {
	ID                 fields.VarUint4         `json:"id"`
	CountryCode        fields.Bytes2           `json:"country_code"`        // 城市
	IdentificationName fields.StringMax255     `json:"identification_name"` // 服务识别码名称
	FeeMin             fields.Amount           `json:"fee_min"`             // 最低手续费
	FeeRatio           fields.VarUint4         `json:"fee_ratio"`           // 最低比率 单位：一亿分之一
	FeeMax             fields.Amount           `json:"fee_max"`             // 最高手续费上限
	Gateway1           fields.StringMax255     `json:"gateway_1"`           // 服务网关域名
	Gateway2           fields.StringMax255     `json:"gateway_2"`           // 备用域名
	OverdueTime        fields.BlockTxTimestamp `json:"overdue_time"`        // 注册服务过期时间
	RegisterTime       fields.BlockTxTimestamp `json:"register_time"`       // 手册注册时间

}

// Is it a legal node name
func IsValidServicerIdentificationName(name string) bool {
	for _, v := range name {
		if false == (v > 'a' && v < 'z' || v > 'A' && v < 'Z' || v > '0' && v < '9') {
			return false
		}
	}
	return true
}

func (m PayRelayNode) Copy() *PayRelayNode {
	newnode := &PayRelayNode{
		ID:                 m.ID,
		CountryCode:        m.CountryCode.Copy(),
		IdentificationName: m.IdentificationName,
		FeeMin:             *m.FeeMin.Copy(),
		FeeRatio:           m.FeeRatio,
		FeeMax:             *m.FeeMax.Copy(),
		Gateway1:           m.Gateway1,
		Gateway2:           m.Gateway2,
		OverdueTime:        m.OverdueTime,
		RegisterTime:       m.RegisterTime,
	}
	return newnode
}

func (m PayRelayNode) Size() uint32 {
	return m.ID.Size() +
		m.CountryCode.Size() +
		m.IdentificationName.Size() +
		m.FeeMin.Size() +
		m.FeeRatio.Size() +
		m.FeeMax.Size() +
		m.Gateway1.Size() +
		m.Gateway2.Size() +
		m.OverdueTime.Size() +
		m.RegisterTime.Size()
}

func (m *PayRelayNode) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.ID.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.CountryCode.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.IdentificationName.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.FeeMin.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.FeeRatio.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.FeeMax.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.Gateway1.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.Gateway2.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.OverdueTime.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.RegisterTime.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m PayRelayNode) Serialize() ([]byte, error) {
	var e error
	var bt []byte = nil
	buf := bytes.NewBuffer(nil)
	bt, e = m.ID.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.CountryCode.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.IdentificationName.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.FeeMin.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.FeeRatio.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.FeeMax.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.Gateway1.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.Gateway2.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.OverdueTime.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.RegisterTime.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	// ok
	return buf.Bytes(), nil
}

// Estimated service charge
func (m PayRelayNode) PredictFeeForPay(payamt *fields.Amount, paysat *fields.Satoshi) (*fields.Amount, *fields.Satoshi) {
	amtfee := fields.NewEmptyAmount()
	satfee := fields.Satoshi(0)
	// Calculate scale
	if !payamt.IsEmpty() {
		bv := payamt.GetValue()
		feeb := new(big.Int).Div(bv, new(big.Int).SetUint64(10000*10000))
		feeb = new(big.Int).Mul(feeb, new(big.Int).SetUint64(uint64(m.FeeRatio)))
		fee, e := fields.NewAmountByBigInt(feeb)
		if e == nil {
			amtfee = fee
		}
	}
	if uint64(*paysat) > 0 {
		satfee = fields.Satoshi(float64(*paysat) * float64(m.FeeRatio) / 100000000)
		if satfee == 0 {
			satfee = 1 // min fee sat = 1
		}
	}
	// Limit height
	if payamt.IsPositive() {
		if m.FeeMin.IsNotEmpty() && amtfee.LessThan(&m.FeeMin) {
			amtfee = m.FeeMin.Copy() // minimum
		} else if m.FeeMax.IsNotEmpty() && amtfee.MoreThan(&m.FeeMax) {
			amtfee = m.FeeMax.Copy() // highest
		}
		// Field size limit
		retfee, _, e := amtfee.CompressForMainNumLen(4, false)
		if e == nil {
			amtfee = retfee
		}
	}
	// ok
	return amtfee, &satfee
}
