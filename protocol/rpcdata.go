package protocol

import (
	"encoding/hex"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hacash/core/fields"
	"io/ioutil"
	"net/http"
)

/**
 * rpc 请求数据
 */

type RpcDataChannelInfo struct {
	//LockBlock    fields.VarUint2 // 单方面结束通道要锁定的区块数量
	ChannelId    fields.ChannelId
	LeftAddress  fields.Address
	LeftAmount   fields.Amount // Mortgage amount 1
	RightAddress fields.Address
	RightAmount  fields.Amount   // Mortgage amount 2
	ReuseVersion fields.VarUint4 // Reuse version number from 1
	Status       fields.VarUint1 // Closed and settled
}

func (r *RpcDataChannelInfo) GetLeftAndRightTotalAmount() *fields.Amount {
	ttamt, _ := r.LeftAmount.Add(&r.RightAmount)
	return ttamt
}

func ParseRpcDataChannelInfoByJSON(cid fields.ChannelId, bytes []byte) (*RpcDataChannelInfo, error) {

	// analysis
	laddr, e := jsonparser.GetString(bytes, "left_address")
	if e != nil && len(laddr) > 0 {
		return nil, fmt.Errorf("Channel not find in rpc data api.")
	}
	leftaddr, _ := fields.CheckReadableAddress(laddr)
	raddr, _ := jsonparser.GetString(bytes, "right_address")
	rightaddr, _ := fields.CheckReadableAddress(raddr)
	status, _ := jsonparser.GetInt(bytes, "status")
	reusev, _ := jsonparser.GetInt(bytes, "reuse_version")
	//lockhei, _ := jsonparser.GetInt(bytes, "lock_block")

	lamt, _ := jsonparser.GetString(bytes, "left_amount")
	lamount, _ := fields.NewAmountFromFinString(lamt)
	ramt, _ := jsonparser.GetString(bytes, "right_amount")
	ramount, _ := fields.NewAmountFromFinString(ramt)

	// return
	channel := &RpcDataChannelInfo{
		//LockBlock:    fields.VarUint2(lockhei),
		ChannelId:    cid,
		ReuseVersion: fields.VarUint4(reusev),
		Status:       fields.VarUint1(status),
		LeftAddress:  *leftaddr,
		LeftAmount:   *lamount,
		RightAddress: *rightaddr,
		RightAmount:  *ramount,
	}
	return channel, nil
}

// Request data from all nodes
func RequestRpcReqChannelInfo(fullNodeRpcUrl string, cid fields.ChannelId) (*RpcDataChannelInfo, error) {

	rurl := fullNodeRpcUrl +
		"/query?action=channel&id=" + hex.EncodeToString(cid)
	resp, e := http.Get(rurl)
	if e != nil {
		return nil, e
	}
	// Request succeeded
	bytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
	}
	//fmt.Println(rurl, string(bytes))
	// Misjudgment
	errmsg, e := jsonparser.GetString(bytes, "errmsg")
	if e == nil && len(errmsg) > 0 {
		return nil, fmt.Errorf(errmsg)
	}
	// analysis
	return ParseRpcDataChannelInfoByJSON(cid, bytes)
}

//////////////////////////////////////////////////////////////

type RpcDataSernodeInfo struct {
	Gateway fields.StringMax255
}

// Request data from all nodes
func RequestChannelAndSernodeInfoFromLoginResolutionApi(apiUrl string, cid fields.ChannelId, sername string) (*RpcDataChannelInfo, *RpcDataSernodeInfo, error) {

	requrl := apiUrl + fmt.Sprintf("/customer/login_resolution?channel_id=%s&servicer_name=%s", hex.EncodeToString(cid), sername)
	//fmt.Println(requrl)
	resp, e := http.Get(requrl)
	if e != nil {
		return nil, nil, e
	}
	// Request succeeded
	bytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, nil, e
	}
	// Misjudgment
	errmsg, e := jsonparser.GetString(bytes, "errmsg")
	if e == nil && len(errmsg) > 0 {
		return nil, nil, fmt.Errorf(errmsg)
	}
	// Resolution 1
	bytes1, _, _, e := jsonparser.Get(bytes, "sernode")
	if e != nil {
		return nil, nil, e
	}
	gateway, e := jsonparser.GetString(bytes1, "gateway")
	if e != nil {
		return nil, nil, fmt.Errorf("Channel not find in rpc data api.")
	}
	var nodeinfo = &RpcDataSernodeInfo{}
	nodeinfo.Gateway = fields.CreateStringMax255(gateway)

	// Resolution 2
	bytes2, _, _, e := jsonparser.Get(bytes, "channel")
	if e != nil {
		return nil, nil, e
	}
	chaninfo, e := ParseRpcDataChannelInfoByJSON(cid, bytes2)
	if e != nil {
		return nil, nil, e
	}

	// success
	return chaninfo, nodeinfo, nil
}
