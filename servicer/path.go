package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/payroutes"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"strings"
)

// Create a single path
func CreatePayPathFormsBySingleNodePath(node *payroutes.PayRelayNode, payamt *fields.Amount, paysat *fields.Satoshi) *protocol.PayPathForms {
	return CreatePayPathForms([][]*payroutes.PayRelayNode{{node}}, payamt, paysat)
}

// Create multiple paths
func CreatePayPathForms(nodepaths [][]*payroutes.PayRelayNode, payamt *fields.Amount, paysat *fields.Satoshi) *protocol.PayPathForms {

	pathdscs := make([]*protocol.PayPathDescribe, len(nodepaths))
	for i, ps := range nodepaths {
		ttfeeamt := fields.NewEmptyAmount()
		ttfeesat := fields.Satoshi(0)
		idlist := make([]fields.VarUint4, len(ps))
		dscs := make([]string, len(ps))
		for n, node := range ps {
			// Handling charge unit: 1000%
			idlist[n] = node.ID
			feeamt, feesat := node.PredictFeeForPay(payamt, paysat)
			feeamtadd, e := ttfeeamt.Add(feeamt)
			feesatadd := ttfeesat + *feesat
			if e == nil {
				ttfeeamt = feeamtadd // Service charge accumulation
			}
			ttfeesat = feesatadd
			dscs[n] = fmt.Sprintf(`%s(fee:%.2f‰)`,
				node.IdentificationName.Value(), float64(node.FeeRatio)/100000)
		}
		dscsall := fmt.Sprintf(`Pay relay: %s`, strings.Join(dscs, " -> "))
		ids := &protocol.NodeIdPath{
			NodeIdCount: fields.VarUint1(len(idlist)),
			NodeIdPath:  idlist,
		}
		paths := &protocol.PayPathDescribe{
			NodeIdPath:        ids,
			PredictPathFeeAmt: *ttfeeamt, // fee
			PredictPathFeeSat: fields.NewSatoshiVariation(uint64(ttfeesat)),
			Describe:          fields.CreateStringMax65535(dscsall),
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
