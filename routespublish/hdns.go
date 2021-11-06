package routespublish

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/x16rs"
	"net/http"
)

/**
 * 用户 HDNS 解析
 */
func (p *PayRoutesPublish) customerAnalyzeHDNS(w http.ResponseWriter, r *http.Request) {

	// diamond
	diastr := protocol.CheckParamString(r, "diamond", "")
	if len(diastr) == 0 {
		protocol.ResponseErrorString(w, "diamond must give.")
		return
	}

	// 判断
	diaok := x16rs.IsDiamondNameOrNumber(diastr)
	if !diaok {
		protocol.ResponseErrorString(w, fmt.Sprintf("<%s> is not a valid diamond name or number.", diastr))
		return
	}

	// 读取钻石归属地址
	apiUrl := p.config.FullNodeRpcURL + "/query?action=hdns&diamond=" + diastr
	realaddr, e := protocol.RequestRpcReqDiamondNameServiceInCommonUse(apiUrl)
	if e != nil {
		protocol.ResponseError(w, e)
		return
	}

	// 成功返回
	data := map[string]interface{}{
		"address": realaddr,
	}
	protocol.ResponseData(w, data)
}
