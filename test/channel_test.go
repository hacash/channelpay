package test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/transactions"
	"github.com/hacashcom/core/account"
	"io"
	"net/http"
	"testing"
)

/*

cd channelpay



rm -rf test/test_node1_data && go build -o test/test_run_node1  ../miner/run/main/main.go && ./test/test_run_node1    testnode1.config.ini

go build -o ./test/test_run_routespublish run/routespublish/main.go && ./test/test_run_routespublish ./testchannelpayroutespublish.config.ini

go build -o ./test/test_run_payservicer1 run/servicer/main.go && ./test/test_run_payservicer1 ./testchannelpayservicer1.config.ini
go build -o ./test/test_run_payservicer2 run/servicer/main.go && ./test/test_run_payservicer2 ./testchannelpayservicer2.config.ini

go build -o ./test/test_run_l2wallet1/w run/client/main.go && ./test/test_run_l2wallet1/w
go build -o ./test/test_run_l2wallet2/w run/client/main.go && ./test/test_run_l2wallet2/w

1Gmc6vNaK9z2SSftEAKsyK2cYvqf2YaXRc_c56934b4a165a0afc129cd93b1130e63_TN1   debugtest1
1G7bhnoo54mdMxQe7dWCGx9kedtNYFihT2_fdfb81f2c55e814b03e1a33653666bc3_TN2   debugtest5

*/

func Test_create_and_commit_test_trs(t *testing.T) {

	/*
		0	1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9 123456
		1	1Gmc6vNaK9z2SSftEAKsyK2cYvqf2YaXRc debugtest1
		2	1GTy2mbWAYFekpfiAn6JRtxyKKnF2TahYa debugtest2
		3	1NjCTuSX8tmYTp68JArrSWX3PsJgN5BrT4 debugtest3
		4	1ATUXj9KoagEKTHe1KRTDdD3UK8rmP4ydr debugtest4
		5	1G7bhnoo54mdMxQe7dWCGx9kedtNYFihT2 debugtest5
		6	1ArBGAe238Wh8BMv41Q9xBjTteTGcZSbXc debugtest6

		1 => 2,3 => 4 => 5
		c56934b4a165a0afc129cd93b1130e63   47295889c6e0e1fc64237e01cd480fd6   fdfb81f2c55e814b03e1a33653666bc3


	*/

	allprikeys := []string{"123456", "debugtest1", "debugtest2", "debugtest3", "debugtest4", "debugtest5", "debugtest6"}
	allPrivateKeyBytes := make(map[string][]byte, len(allprikeys))
	alltestacc := make([]*account.Account, len(allprikeys))
	for i, v := range allprikeys {
		acc := account.GetAccountByPrivateKeyOrPassword(v)
		allPrivateKeyBytes[string(acc.Address)] = acc.PrivateKey
		alltestacc[i] = acc
	}

	var hx1, _ = hex.DecodeString("00000000006ff7e57530f44b3326b1edc56934b4a165a0afc129cd93b1130e63")
	var mainacc = account.GetAccountByPrivateKeyOrPassword("123456")

	var trspkg = transactions.Transaction_2_Simple{
		Timestamp:   1691642059,
		MainAddress: mainacc.Address,
		Fee:         *fields.NewAmountNumSmallCoin(1),
	}
	// chain ID = 1
	trspkg.AddAction(&actions.Action_30_SupportDistinguishForkChainID{CheckChainID: 1})

	// btc move
	trspkg.AddAction(&actions.Action_7_SatoshiGenesis{
		TransferNo:               1,
		BitcoinBlockHeight:       1,
		BitcoinBlockTimestamp:    1691642059,
		BitcoinEffectiveGenesis:  0,
		BitcoinQuantity:          1023,
		AdditionalTotalHacAmount: 10485760,
		OriginAddress:            mainacc.Address,
		BitcoinTransferHash:      hx1,
	})

	// transfer BTC
	for i, v := range alltestacc {
		if i == 0 {
			continue
		} // over main
		trspkg.AddAction(&actions.Action_8_SimpleSatoshiTransfer{
			ToAddress: v.Address,
			Amount:    50 * 100000000,
		})
	}

	// transfer HAC
	for i, v := range alltestacc {
		if i == 0 {
			continue
		} // over main
		trspkg.AddAction(&actions.Action_1_SimpleToTransfer{
			ToAddress: v.Address,
			Amount:    *fields.NewAmountNumSmallCoin(50),
		})
	}

	// Open channel
	// channel 1
	channel_id_1, _ := hex.DecodeString("c56934b4a165a0afc129cd93b1130e63")
	channel_id_2, _ := hex.DecodeString("47295889c6e0e1fc64237e01cd480fd6")
	channel_id_3, _ := hex.DecodeString("fdfb81f2c55e814b03e1a33653666bc3")
	trspkg.AddAction(&actions.Action_31_OpenPaymentChannelWithSatoshi{
		ChannelId:            channel_id_1,
		ArbitrationLockBlock: 30,
		InterestAttribution:  0,
		LeftAddress:          alltestacc[1].Address,
		LeftAmount:           *fields.NewAmountNumSmallCoin(10),
		LeftSatoshi:          fields.NewSatoshiVariation(10 * 100000000),
		RightAddress:         alltestacc[2].Address,
		RightAmount:          *fields.NewAmountNumSmallCoin(10),
		RightSatoshi:         fields.NewSatoshiVariation(10 * 100000000),
	})
	trspkg.AddAction(&actions.Action_31_OpenPaymentChannelWithSatoshi{
		ChannelId:            channel_id_2,
		ArbitrationLockBlock: 30,
		InterestAttribution:  0,
		LeftAddress:          alltestacc[3].Address,
		LeftAmount:           *fields.NewAmountNumSmallCoin(10),
		LeftSatoshi:          fields.NewSatoshiVariation(10 * 100000000),
		RightAddress:         alltestacc[4].Address,
		RightAmount:          *fields.NewAmountNumSmallCoin(10),
		RightSatoshi:         fields.NewSatoshiVariation(10 * 100000000),
	})
	trspkg.AddAction(&actions.Action_31_OpenPaymentChannelWithSatoshi{
		ChannelId:            channel_id_3,
		ArbitrationLockBlock: 30,
		InterestAttribution:  0,
		LeftAddress:          alltestacc[4].Address,
		LeftAmount:           *fields.NewAmountNumSmallCoin(10),
		LeftSatoshi:          fields.NewSatoshiVariation(10 * 100000000),
		RightAddress:         alltestacc[5].Address,
		RightAmount:          *fields.NewAmountNumSmallCoin(10),
		RightSatoshi:         fields.NewSatoshiVariation(10 * 100000000),
	})

	/********************************8*/

	//do sign
	trspkg.FillNeedSigns(allPrivateKeyBytes, nil)

	// commit trs
	trsbytes, _ := trspkg.Serialize()
	//isonlytry := true
	isonlytry := false
	err := commitTrs("127.0.0.1:8087", trsbytes, isonlytry)
	fmt.Println(err)

}

func commitTrs(domain string, txbody []byte, isonlytry bool) error {
	var url = fmt.Sprintf("http://%s/submit?action=transaction", domain)
	if isonlytry {
		url += "&onlytry=true"
	}
	fmt.Println(url)
	var req, e = http.NewRequest("POST", url, bytes.NewReader(txbody))
	if e != nil {
		return e
	}
	var cli = &http.Client{}
	res, e := cli.Do(req)
	if e != nil {
		return e
	}
	defer res.Body.Close()

	bdbts, e := io.ReadAll(res.Body)
	if e != nil {
		return e
	}

	// ok
	fmt.Println(string(bdbts))
	return nil
}

func Test_t2(t *testing.T) {

}
