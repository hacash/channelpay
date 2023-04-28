package protocol

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"regexp"
	"strings"
)

/**
 * 通道链地址解析
 */

type ChannelAccountAddress struct {
	Address      fields.Address      // address
	ChannelId    fields.ChannelId    // Channel Chain ID
	ServicerName fields.StringMax255 // Name of service provider

}

// Resolve new address
func ParseChannelAccountAddress(addrstr string) (*ChannelAccountAddress, error) {
	addr := &ChannelAccountAddress{}
	e := addr.Parse(addrstr)
	if e != nil {
		return nil, e
	}
	return addr, nil
}

// Comparison operator name
func (c *ChannelAccountAddress) CompareServiceName(sname string) bool {
	if strings.Compare(
		strings.ToLower(c.ServicerName.Value()),
		strings.ToLower(sname),
	) == 0 {
		return true
	}
	// Different
	return false
}

// Readable address
func (c *ChannelAccountAddress) ToReadable(isstrict bool) string {
	addr := c.Address.ToReadable()
	if isstrict && len(c.ChannelId) == stores.ChannelIdLength {
		addr += "_" + c.ChannelId.ToHex()
	}
	addr += "_" + c.ServicerName.Value()
	return addr
}

// Resolve address
func (c *ChannelAccountAddress) Parse(addrstr string) error {
	addrstr1 := regexp.MustCompile(`^_+|_+$`).ReplaceAllString(addrstr, "")
	addrstr2 := regexp.MustCompile(`_+`).ReplaceAllString(addrstr1, "_")
	addrObj := strings.Split(addrstr2, "_")
	if len(addrObj) == 0 {
		return fmt.Errorf("channel address cannot be empty.")
	}
	sAddr := addrObj[0]
	channelId := ""
	serName := ""
	if len(addrObj) == 2 {
		serName = addrObj[1]
	} else if len(addrObj) == 3 {
		channelId = addrObj[1]
		serName = addrObj[2]
	} else {
		return fmt.Errorf("%s is not a valid channel address.", addrstr)
	}
	// addr
	addr, e := fields.CheckReadableAddress(sAddr)
	if e != nil {
		return fmt.Errorf("address <%s> is not valid", sAddr)
	}
	c.Address = *addr
	// cid
	if len(channelId) > 0 {
		cidbts, e := hex.DecodeString(channelId)
		if e != nil || len(cidbts) != stores.ChannelIdLength {
			return fmt.Errorf("channel id <%s> is not valid", channelId)
		}
		c.ChannelId = cidbts
	}
	// service name
	if len(serName) == 0 ||
		len(regexp.MustCompile(`[A-Za-z0-9]+`).ReplaceAllString(serName, "")) != 0 {
		return fmt.Errorf("service name <%s> is not valid", serName)
	}
	c.ServicerName = fields.CreateStringMax255(serName)

	// Resolution succeeded
	return nil
}
