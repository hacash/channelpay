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

// Determine whether it is the hdns diamond domain name service address
func IsHDNSaddress(addrstr string) (string, bool) {
	addrObj := strings.Split(addrstr, "_")
	if len(addrObj) < 1 {
		return "", false
	}
	diakind := addrObj[0]
	if x16rs.IsDiamondValueString(diakind) {
		return diakind, true // Diamond face value
	}
	if num, ok := strconv.Atoi(diakind); ok == nil && num > 0 && num < 16777216 {
		return diakind, true // Diamond number
	}
	return "", false
}

// Request hdns resolution data from all nodes
func RequestRpcReqDiamondNameServiceFromLoginResolutionApi(apiDomain string, diakind string) (string, error) {
	// Hdns service
	requrl := apiDomain + fmt.Sprintf("/customer/hdns_analyze?diamond=%s", diakind)
	//fmt.Println(requrl)
	return RequestRpcReqDiamondNameServiceInCommonUse(requrl)
}

// Request hdns resolution data from all nodes
func RequestRpcReqDiamondNameServiceInCommonUse(apiUrl string) (string, error) {
	// Hdns service
	//fmt.Println(requrl)
	resp, e := http.Get(apiUrl)
	if e != nil {
		return "", e
	}
	// Request succeeded
	bytes, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return "", e
	}
	// Misjudgment
	errmsg, e := jsonparser.GetString(bytes, "errmsg")
	if e == nil && len(errmsg) > 0 {
		return "", fmt.Errorf(errmsg)
	}
	// To address
	address, e := jsonparser.GetString(bytes, "address")
	if e != nil {
		return "", e
	}
	// Resolution succeeded
	return address, nil
}
