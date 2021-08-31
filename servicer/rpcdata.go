package servicer

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
	LockBlock    fields.VarUint2 // 单方面结束通道要锁定的区块数量
	LeftAddress  fields.Address
	LeftAmount   fields.Amount // 抵押数额1
	RightAddress fields.Address
	RightAmount  fields.Amount   // 抵押数额2
	ReuseVersion fields.VarUint4 // 重用版本号 从 1 开始
	Status       fields.VarUint1 // 已经关闭并结算等状态
}

func (s *Servicer) rpcReqChannelInfo(cid fields.Bytes16) (*RpcDataChannelInfo, error) {

	resp, e := http.Get(s.config.FullNodeRpcUrl +
		"/query?action=channel&id=" + hex.EncodeToString(cid))
	if e != nil {
		return nil, e
	}
	// 请求成功
	bytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
	}
	// 解析
	laddr, e := jsonparser.GetString(bytes, "left_address")
	if e != nil && len(laddr) > 0 {
		fmt.Errorf("Channel not find in rpc data api.")
	}
	leftaddr, _ := fields.CheckReadableAddress(laddr)
	raddr, _ := jsonparser.GetString(bytes, "right_address")
	rightaddr, _ := fields.CheckReadableAddress(raddr)
	status, _ := jsonparser.GetInt(bytes, "status")
	reusev, _ := jsonparser.GetInt(bytes, "reuse_version")
	lockhei, _ := jsonparser.GetInt(bytes, "lock_block")

	lamt, _ := jsonparser.GetString(bytes, "left_amount")
	lamount, _ := fields.NewAmountFromFinString(lamt)
	ramt, _ := jsonparser.GetString(bytes, "right_amount")
	ramount, _ := fields.NewAmountFromFinString(ramt)

	// 返回
	channel := &RpcDataChannelInfo{
		LockBlock:    fields.VarUint2(lockhei),
		ReuseVersion: fields.VarUint4(reusev),
		Status:       fields.VarUint1(status),
		LeftAddress:  *leftaddr,
		LeftAmount:   *lamount,
		RightAddress: *rightaddr,
		RightAmount:  *ramount,
	}

	return channel, nil

}
