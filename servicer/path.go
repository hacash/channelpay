package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/payroutes"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"strings"
)

// 创建单个路径
func CreatePayPathFormsBySingleNodePath(node *payroutes.PayRelayNode, payamt *fields.Amount) *protocol.PayPathForms {
	return CreatePayPathForms([][]*payroutes.PayRelayNode{{node}}, payamt)
}

// 创建多条路径
func CreatePayPathForms(nodepaths [][]*payroutes.PayRelayNode, payamt *fields.Amount) *protocol.PayPathForms {

	pathdscs := make([]*protocol.PayPathDescribe, len(nodepaths))
	for i, ps := range nodepaths {
		ttfee := fields.NewEmptyAmount()
		idlist := make([]fields.VarUint4, len(ps))
		dscs := make([]string, len(ps))
		for n, node := range ps {
			// 手续费单位：千分之
			idlist[n] = node.ID
			fee := node.PredictFeeForPay(payamt)
			feeadd, e := ttfee.Add(fee)
			if e == nil {
				ttfee = feeadd // 手续费累加
			}
			dscs[n] = fmt.Sprintf(`%s(fee:%.2f‰)`,
				node.IdentificationName, float64(node.FeeRatio)/100000)
		}
		dscsall := fmt.Sprintf(`Pay relay: %s`, strings.Join(dscs, " -> "))
		ids := &protocol.NodeIdPath{
			NodeIdCount: fields.VarUint1(len(idlist)),
			NodeIdPath:  idlist,
		}
		paths := &protocol.PayPathDescribe{
			NodeIdPath:     ids,
			PredictPathFee: *ttfee, // 手续费
			Describe:       fields.CreateStringMax65535(dscsall),
		}
		pathdscs[i] = paths

	}
	// 最终
	forms := &protocol.PayPathForms{
		PayPathCount: fields.VarUint1(len(nodepaths)),
		PayPaths:     pathdscs,
	}
	// 完毕
	return forms
}
