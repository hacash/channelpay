package servicer

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/stores"
	"strings"
)

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
