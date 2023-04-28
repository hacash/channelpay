package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/payroutes"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"strings"
)

// Create a single path
func CreatePayPathFormsBySingleNodePath(node *payroutes.PayRelayNode, payamt *fields.Amount) *protocol.PayPathForms {
	return CreatePayPathForms([][]*payroutes.PayRelayNode{{node}}, payamt)
}

// Create multiple paths
func CreatePayPathForms(nodepaths [][]*payroutes.PayRelayNode, payamt *fields.Amount) *protocol.PayPathForms {

	pathdscs := make([]*protocol.PayPathDescribe, len(nodepaths))
	for i, ps := range nodepaths {
		ttfee := fields.NewEmptyAmount()
		idlist := make([]fields.VarUint4, len(ps))
		dscs := make([]string, len(ps))
		for n, node := range ps {
			// Handling charge unit: 1000%
			idlist[n] = node.ID
			fee := node.PredictFeeForPay(payamt)
			feeadd, e := ttfee.Add(fee)
			if e == nil {
				ttfee = feeadd // Service charge accumulation
			}
			dscs[n] = fmt.Sprintf(`%s(fee:%.2f‰)`,
				node.IdentificationName.Value(), float64(node.FeeRatio)/100000)
		}
		dscsall := fmt.Sprintf(`Pay relay: %s`, strings.Join(dscs, " -> "))
		ids := &protocol.NodeIdPath{
			NodeIdCount: fields.VarUint1(len(idlist)),
			NodeIdPath:  idlist,
		}
		paths := &protocol.PayPathDescribe{
			NodeIdPath:     ids,
			PredictPathFee: *ttfee, // 预估手续费
			Describe:       fields.CreateStringMax65535(dscsall),
		}
		pathdscs[i] = paths

	}
	// final
	forms := &protocol.PayPathForms{
		PayPathCount: fields.VarUint1(len(nodepaths)),
		PayPaths:     pathdscs,
	}
	// complete
	return forms
}
