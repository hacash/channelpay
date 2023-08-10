package routespublish

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/stores"
	"net/http"
)

/**
 * 用户登录解析
 */
func (p *PayRoutesPublish) customerLoginResolution(w http.ResponseWriter, r *http.Request) {

	// Channel ID
	cidstr := protocol.CheckParamString(r, "channel_id", "")
	if len(cidstr) == 0 {
		protocol.ResponseErrorString(w, "channel_id must give.")
		return
	}
	cidbt, e := hex.DecodeString(cidstr)
	if e != nil || len(cidbt) != stores.ChannelIdLength {
		protocol.ResponseErrorString(w, "channel_id format error.")
		return
	}

	// Name of payment service provider
	svername := protocol.CheckParamString(r, "servicer_name", "")
	if len(svername) == 0 {
		protocol.ResponseErrorString(w, "servicer_name must give.")
		return
	}

	// Find whether the service provider exists
	snode := p.routingManager.FindNodeByName(svername)
	if snode == nil {
		protocol.ResponseError(w, fmt.Errorf("servicer %s not find in routes list.", svername))
		return
	}

	// Read channel status
	channelInfo, e := protocol.RequestRpcReqChannelInfo(p.config.FullNodeRpcURL, cidbt)
	if e != nil {
		protocol.ResponseError(w, fmt.Errorf("request channel info error: ", e.Error()))
		return
	}

	// Check channel status
	if channelInfo.Status != stores.ChannelStatusOpening {
		protocol.ResponseError(w, fmt.Errorf("channel status is not on opening!"))
		return
	}

	// return
	var nodeinfo = make(map[string]interface{})
	var chaninfo = make(map[string]interface{})
	// node
	nodeinfo["gateway"] = snode.Gateway1.Value()
	// channel
	chaninfo["status"] = channelInfo.Status
	chaninfo["reuse_version"] = channelInfo.ReuseVersion
	chaninfo["left_address"] = channelInfo.LeftAddress.ToReadable()
	chaninfo["left_amount"] = channelInfo.LeftAmount.ToFinString()
	chaninfo["left_satoshi"] = channelInfo.LeftSatoshi
	chaninfo["right_address"] = channelInfo.RightAddress.ToReadable()
	chaninfo["right_amount"] = channelInfo.RightAmount.ToFinString()
	chaninfo["right_satoshi"] = channelInfo.RightSatoshi

	// Return resolution
	protocol.ResponseData(w, map[string]interface{}{
		"sernode": nodeinfo,
		"channel": chaninfo,
	})

}
