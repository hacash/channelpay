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
	LeftAddress  fields.Address
	LeftAmount   fields.Amount // 抵押数额1
	RightAddress fields.Address
	RightAmount  fields.Amount   // 抵押数额2
	ReuseVersion fields.VarUint4 // 重用版本号 从 1 开始
	Status       fields.VarUint1 // 已经关闭并结算等状态
}

func ParseRpcDataChannelInfoByJSON(bytes []byte) (*RpcDataChannelInfo, error) {

	// 解析
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

	// 返回
	channel := &RpcDataChannelInfo{
		//LockBlock:    fields.VarUint2(lockhei),
		ReuseVersion: fields.VarUint4(reusev),
		Status:       fields.VarUint1(status),
		LeftAddress:  *leftaddr,
		LeftAmount:   *lamount,
		RightAddress: *rightaddr,
		RightAmount:  *ramount,
	}
	return channel, nil
}

// 向全节点请求数据
func RequestRpcReqChannelInfo(fullNodeRpcUrl string, cid fields.ChannelId) (*RpcDataChannelInfo, error) {

	rurl := fullNodeRpcUrl +
		"/query?action=channel&id=" + hex.EncodeToString(cid)
	resp, e := http.Get(rurl)
	if e != nil {
		return nil, e
	}
	// 请求成功
	bytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
	}
	//fmt.Println(rurl, string(bytes))
	// 判断错误
	errmsg, e := jsonparser.GetString(bytes, "errmsg")
	if e == nil && len(errmsg) > 0 {
		return nil, fmt.Errorf(errmsg)
	}
	// 解析
	return ParseRpcDataChannelInfoByJSON(bytes)
}

//////////////////////////////////////////////////////////////

type RpcDataSernodeInfo struct {
	Gateway fields.StringMax255
}

// 向全节点请求数据
func RequestChannelAndSernodeInfoFromLoginResolutionApi(apiUrl string, cid fields.ChannelId, sername string) (*RpcDataChannelInfo, *RpcDataSernodeInfo, error) {

	requrl := apiUrl + fmt.Sprintf("/customer/login_resolution?channel_id=%s&servicer_name=%s", hex.EncodeToString(cid), sername)
	//fmt.Println(requrl)
	resp, e := http.Get(requrl)
	if e != nil {
		return nil, nil, e
	}
	// 请求成功
	bytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, nil, e
	}
	// 判断错误
	errmsg, e := jsonparser.GetString(bytes, "errmsg")
	if e == nil && len(errmsg) > 0 {
		return nil, nil, fmt.Errorf(errmsg)
	}
	// 解析1
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

	// 解析2
	bytes2, _, _, e := jsonparser.Get(bytes, "channel")
	if e != nil {
		return nil, nil, e
	}
	chaninfo, e := ParseRpcDataChannelInfoByJSON(bytes2)
	if e != nil {
		return nil, nil, e
	}

	// 成功
	return chaninfo, nodeinfo, nil
}
