package servicer

import (
	"encoding/hex"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/payroutes"
	"github.com/hacash/core/stores"
	"github.com/hacash/core/sys"
	"io/ioutil"
	"path"
	"strings"
)

// 初始化结算通道
func (s *Servicer) setupRelaySettlementChannelDataSettings() {
	var jsonfilecon []byte = []byte(s.config.RelaySettlementChannelsJsonFile)
	if path.Ext(s.config.RelaySettlementChannelsJsonFile) == ".json" {
		// 读取文件
		fpth := sys.AbsDir(s.config.RelaySettlementChannelsJsonFile)
		bts, e := ioutil.ReadFile(fpth)
		if e == nil {
			jsonfilecon = bts
		}
	}
	s.settlenoderChgLock.Lock()
	defer s.settlenoderChgLock.Unlock()
	// 解析json
	jsonparser.ObjectEach(jsonfilecon, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		k := string(key)
		if payroutes.IsValidServicerIdentificationName(k) {
			if s.settlenoder[k] == nil {
				s.settlenoder[k] = make([]*chanpay.RelayPaySettleNoder, 0)
			}
			// 循环多个通道
			jsonparser.ArrayEach(value, func(item []byte, dataType jsonparser.ValueType, offset int, err error) {
				if len(item) == 32+2 {
					cid, _ := hex.DecodeString(string(item[2:]))
					if len(cid) != stores.ChannelIdLength {
						return // 错误
					}
					weLeft := true
					if item[0] == 'r' {
						weLeft = false
					}
					node := chanpay.NewRelayPayNodeConnect(k, cid, weLeft, nil)
					// 添加一个通道
					s.settlenoder[k] = append(s.settlenoder[k], node)
				}
			})
		}
		return nil
	})

}

// 修改通道数据设定
func (s *Servicer) modifyChannelDataSettings() {
	// 删除客户服务通道
	sccs := strings.Split(strings.Replace(s.config.ServiceCustomerChannelsCancel, " ", "", -1), ",")
	sccssn := 0
	for _, v := range sccs {
		cid, e := hex.DecodeString(v)
		if e == nil && len(cid) == stores.ChannelIdLength {
			s.chanset.CancelCustomerPayChannel(cid) // 取消服务
			sccssn++
		}
	}
	if sccssn > 0 {
		fmt.Printf("[Config] ServiceCustomerChannels Cancel %d channels.\n", sccssn)
	}

	// 添加客户服务通道
	scas := strings.Split(strings.Replace(s.config.ServiceCustomerChannelsAdd, " ", "", -1), ",")
	scassn := 0
	for _, v := range scas {
		cid, e := hex.DecodeString(v)
		if e == nil && len(cid) == stores.ChannelIdLength {
			s.chanset.SetupCustomerPayChannel(cid) // 添加服务
			scassn++
		}
	}
	if scassn > 0 {
		fmt.Printf("[Config] ServiceCustomerChannels Add %d channels.\n", scassn)
	}

}