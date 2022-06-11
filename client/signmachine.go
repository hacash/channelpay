package client

import (
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
)

/**
 * 签名机
 */

// Signature machine
type SignatureMachine struct {
	acc *account.Account
}

func NewSignatureMachine(acc *account.Account) *SignatureMachine {
	return &SignatureMachine{
		acc: acc,
	}
}

func (s *SignatureMachine) Init() error {
	return nil
}
func (s *SignatureMachine) TemporaryStoragePrivateKeyForSign(string) {}
func (s *SignatureMachine) RemovePrivateKey(fields.Address)          {}
func (s *SignatureMachine) CleanAllPrivateKey()                      {}

// Sign the statement and then check all signatures
func (s *SignatureMachine) CheckReconciliationFillNeedSignature(bill *channel.OffChainFormPaymentChannelRealtimeReconciliation, checksign *fields.Sign) (*fields.Sign, error) {
	return nil, nil
}

// Send the channel transaction to the signer to verify the data, and automatically fill in the signature
func (s *SignatureMachine) CheckPaydocumentAndFillNeedSignature(paydocs *channel.ChannelPayCompleteDocuments, mustaddrs []fields.Address) (*fields.SignListMax255, error) {
	// Check account
	if len(mustaddrs) != 1 {
		return nil, fmt.Errorf("mustaddrs length error")
	}
	if mustaddrs[0].NotEqual(s.acc.Address) {
		return nil, fmt.Errorf("mustaddrs error, need address %s but got %s",
			s.acc.AddressReadable, mustaddrs[0].ToReadable())
	}
	// Whether the downstream has completed the signature has been checked before calling this function

	// Execute signature
	sghx := paydocs.ChainPayment.GetSignStuffHash()
	signature, e := s.acc.Private.Sign(sghx)
	if e != nil {
		return nil, fmt.Errorf("do sign error: %s", e.Error())
	}
	signobj := fields.Sign{
		PublicKey: s.acc.PublicKey,
		Signature: signature.Serialize64(),
	}

	// Return signature
	return &fields.SignListMax255{
		Count: 1,
		Signs: []fields.Sign{signobj},
	}, nil
}
