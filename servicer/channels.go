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
	"os"
	"path"
	"strings"
)

// Initialize password
func (s *Servicer) setupPasswordSettings() {
	var passwordstr string = s.config.SignatureMachinePrivateKeySetupList
	if path.Ext(s.config.SignatureMachinePrivateKeySetupList) == ".txt" {
		// read file
		fpth := sys.AbsDir(s.config.SignatureMachinePrivateKeySetupList)
		bts, e := ioutil.ReadFile(fpth)
		if e == nil {
			passwordstr = string(bts)
			//fmt.Println(passwordstr, "**************************")
		}
	}
	s.settlenoderChgLock.Lock()
	defer s.settlenoderChgLock.Unlock()
	// Take out line breaks and spaces
	passwordstr = strings.Replace(passwordstr, " ", "", -1)
	//passwordstr = strings.Replace(passwordstr, "\n", "", -1)
	// Parse password
	//fmt.Println(passwordstr)
	for _, v := range strings.Split(passwordstr, "\n") {
		if len(v) < 6 {
			continue
		}
		s.signmachine.TemporaryStoragePrivateKeyForSign(v)
	}
}

// Initialize settlement channel
func (s *Servicer) setupRelaySettlementChannelDataSettings() {
	var jsonfilecon []byte = []byte(s.config.RelaySettlementChannelsJsonFile)
	if path.Ext(s.config.RelaySettlementChannelsJsonFile) == ".json" {
		// read file
		fpth := sys.AbsDir(s.config.RelaySettlementChannelsJsonFile)
		bts, e := ioutil.ReadFile(fpth)
		if e == nil {
			jsonfilecon = bts
		}
	}
	s.settlenoderChgLock.Lock()
	defer s.settlenoderChgLock.Unlock()
	// Parsing JSON
	var chnumcount = 0
	fmt.Printf("[InitializeChannelSide]")
	jsonparser.ObjectEach(jsonfilecon, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		k := string(key)
		if payroutes.IsValidServicerIdentificationName(k) {
			if s.settlenoder[k] == nil {
				s.settlenoder[k] = make([]*chanpay.RelayPaySettleNoder, 0)
			}
			fmt.Printf(" " + k + ":")
			// Cycle multiple channels
			jsonparser.ArrayEach(value, func(item []byte, dataType jsonparser.ValueType, offset int, err error) {
				if len(item) == 32+2 {
					cidstr := string(item[2:])
					cid, _ := hex.DecodeString(cidstr)
					if len(cid) != stores.ChannelIdLength {
						return // error
					}
					weLeft := true
					if item[0] == 'r' {
						weLeft = false
					}
					// New channel
					sise := chanpay.NewChannelSideById(cid)
					e := s.InitializeChannelSide(sise, nil, weLeft)
					if e != nil {
						// Error initializing settlement channel
						fmt.Printf("InitializeChannelSide %s error:\n", cidstr)
						fmt.Println(e.Error())
						os.Exit(0)
						return
					}
					fmt.Printf(" %s", cidstr)
					node := chanpay.NewRelayPayNodeConnect(k, cid, weLeft, sise)
					// Add a channel
					s.settlenoder[k] = append(s.settlenoder[k], node)
					chnumcount++
				}
			})
		}
		fmt.Printf(" all %d ok.\n", chnumcount)
		return nil
	})

}

// Modify channel data settings
func (s *Servicer) modifyChannelDataSettings() {
	// Delete customer service channel
	sccs := strings.Split(strings.Replace(s.config.ServiceCustomerChannelsCancel, " ", "", -1), ",")
	sccssn := 0
	for _, v := range sccs {
		cid, e := hex.DecodeString(v)
		if e == nil && len(cid) == stores.ChannelIdLength {
			s.chanset.CancelCustomerPayChannel(cid) // Cancel service
			sccssn++
		}
	}
	if sccssn > 0 {
		fmt.Printf("[Config] ServiceCustomerChannels Cancel %d channels.\n", sccssn)
	}

	// Add customer service channel
	scas := strings.Split(strings.Replace(s.config.ServiceCustomerChannelsAdd, " ", "", -1), ",")
	scassn := 0
	for _, v := range scas {
		cid, e := hex.DecodeString(v)
		if e == nil && len(cid) == stores.ChannelIdLength {
			s.chanset.SetupCustomerPayChannel(cid) // Add service
			scassn++
		}
	}
	if scassn > 0 {
		fmt.Printf("[Config] ServiceCustomerChannels Add %d channels.\n", scassn)
	}

}
