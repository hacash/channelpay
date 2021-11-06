package protocol

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hacash/x16rs"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// 判断是否为 HDNS 钻石域名服务地址
func IsHDNSaddress(addrstr string) (string, bool) {
	addrObj := strings.Split(addrstr, "_")
	if len(addrObj) < 1 {
		return "", false
	}
	diakind := addrObj[0]
	if x16rs.IsDiamondValueString(diakind) {
		return diakind, true // 钻石字面值
	}
	if num, ok := strconv.Atoi(diakind); ok == nil && num > 0 && num < 16777216 {
		return diakind, true // 钻石编号
	}
	return "", false
}

// 向全节点请求 HDNS 解析数据
func RequestRpcReqDiamondNameServiceFromLoginResolutionApi(apiDomain string, diakind string) (string, error) {
	// hdns 服务
	requrl := apiDomain + fmt.Sprintf("/customer/hdns_analyze?diamond=%s", diakind)
	//fmt.Println(requrl)
	return RequestRpcReqDiamondNameServiceInCommonUse(requrl)
}

// 向全节点请求 HDNS 解析数据
func RequestRpcReqDiamondNameServiceInCommonUse(apiUrl string) (string, error) {
	// hdns 服务
	//fmt.Println(requrl)
	resp, e := http.Get(apiUrl)
	if e != nil {
		return "", e
	}
	// 请求成功
	bytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return "", e
	}
	// 判断错误
	errmsg, e := jsonparser.GetString(bytes, "errmsg")
	if e == nil && len(errmsg) > 0 {
		return "", fmt.Errorf(errmsg)
	}
	// 去地址
	address, e := jsonparser.GetString(bytes, "address")
	if e != nil {
		return "", e
	}
	// 解析成功
	return address, nil
}
