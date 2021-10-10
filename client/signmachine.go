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

// 签名机
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

// 将通道交易送入签名机验证数据，并自动填充签名
func (s *SignatureMachine) CheckPaydocumentAndFillNeedSignature(paydocs *channel.ChannelPayCompleteDocuments, mustaddrs []fields.Address) (*fields.SignListMax255, error) {
	// 检查账户
	if len(mustaddrs) != 1 {
		return nil, fmt.Errorf("mustaddrs length error")
	}
	if mustaddrs[0].NotEqual(s.acc.Address) {
		return nil, fmt.Errorf("mustaddrs error, need address %s but got %s",
			s.acc.AddressReadable, mustaddrs[0].ToReadable())
	}
	// 下游是否完成签名，已经在调用本函数前完成检查

	// 执行签名
	sghx := paydocs.ChainPayment.GetSignStuffHash()
	signature, e := s.acc.Private.Sign(sghx)
	if e != nil {
		return nil, fmt.Errorf("do sign error: %s", e.Error())
	}
	signobj := fields.Sign{
		PublicKey: s.acc.PublicKey,
		Signature: signature.Serialize64(),
	}

	// 返回签名
	return &fields.SignListMax255{
		Count: 1,
		Signs: []fields.Sign{signobj},
	}, nil
}
