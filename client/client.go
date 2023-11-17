package client

import (
	"encoding/hex"
	"fmt"
	fyne "fyne.io/fyne/v2"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/zserge/lorca"
	"log"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// Structure to be paid
type pendingPayment struct {
	address     protocol.ChannelAccountAddress
	amount      fields.Amount
	satoshi     fields.SatoshiVariation
	prequeryMsg *protocol.MsgResponsePrequeryPayment
}

/**
 * 通道链支付客户端
 */
type ChannelPayClient struct {
	app         fyne.App
	loginWindow fyne.Window
	payui       lorca.UI
	//window fyne.Window
	user *ChannelPayUser // Client
	// Pending payment cache data
	pendingPaymentObj *pendingPayment

	// Status lock
	statusMutex sync.Mutex
}

func CreateChannelPayClient(app fyne.App, user *ChannelPayUser, lgwd fyne.Window) *ChannelPayClient {
	return &ChannelPayClient{
		app:               app,
		loginWindow:       lgwd,
		user:              user,
		pendingPaymentObj: nil,
	}
}

// Update interface display
func (c *ChannelPayClient) UpdateBalanceShow() {
	cside := c.user.servicerStreamSide.ChannelSide
	bill := cside.GetReconciliationBill()
	billhex := ""
	if bill != nil {
		bts, _ := bill.SerializeWithTypeCode()
		billhex = hex.EncodeToString(bts)
	}
	c.payui.Eval(fmt.Sprintf(`UpdateBalance("%s","%s",%d,%d,%d,%d,"%s")`,
		cside.GetChannelCapacityAmountOfOur().ToFinString(),
		cside.GetChannelCapacityAmountOfRemote().ToFinString(),
		cside.GetChannelCapacitySatoshiOfOur(),
		cside.GetChannelCapacitySatoshiOfRemote(),
		cside.GetAvailableReuseVersion(),
		cside.GetAvailableAutoNumber(),
		billhex,
	))
}

// Start transaction
func (c *ChannelPayClient) BindFuncPrequeryPayment(addr, amt string) string {
	//fmt.Println("BindFuncInitiatePayment:", addr, amt)
	addr = strings.Trim(addr, "\n ")
	amt = strings.ToUpper(strings.Trim(amt, "\n "))
	if len(addr) == 0 {
		return "Please enter the address."
	}
	if len(amt) == 0 {
		return "Please enter the amount."
	}
	// Resolve hdns
	diakind, ok := protocol.IsHDNSaddress(addr)
	if ok {
		// Perform resolution
		apiurl := GetLoginResolutionApiDomain()
		realaddr, err := protocol.RequestRpcReqDiamondNameServiceFromLoginResolutionApi(apiurl, diakind)
		if err != nil {
			return fmt.Sprintf("Address Diamond Name Service error: %s", err.Error()) // 错误
		}
		addrary := strings.Split(addr, "_")
		addrary[0] = realaddr
		addr = strings.Join(addrary, "_")
		// Parsing log printing
		c.ShowStatusLog(fmt.Sprintf("HDNS analyze: diamond(%s) => %s", diakind, realaddr))
	}
	acc, e := protocol.ParseChannelAccountAddress(addr)
	if e != nil {
		return fmt.Sprintf("Address format error: %s", e.Error()) // 错误
	}
	amount, e := fields.NewAmountFromString(amt)
	satoshi := fields.NewSatoshiVariation(0)
	var sts uint64 = 0
	if strings.Contains(amt, "SAT") {
		sts, e = strconv.ParseUint(strings.Trim(amt, "SAT"), 10, 64)
		if e == nil && sts > 0 {
			satoshi = fields.NewSatoshiVariation(sts)
		}
	}
	if sts == 0 && amount == nil {
		return fmt.Sprintf("Amount format error: %s", e.Error()) // 错误
	}
	if amount == nil {
		amount = fields.NewEmptyAmount()
	}
	// Balance check
	// check SAT
	stscap := c.user.servicerStreamSide.ChannelSide.GetChannelCapacitySatoshiOfOur()
	if uint64(stscap) < sts {
		return fmt.Sprintf("Balance %d sats not enough for transfer %d sts",
			stscap, sts) // Sorry, your credit is running low
	}
	// check HAC
	amtcap := c.user.servicerStreamSide.ChannelSide.GetChannelCapacityAmountOfOur()
	if amtcap.LessThan(amount) {
		return fmt.Sprintf("Balance %s not enough for transfer %s",
			amtcap.ToFinString(), amount.ToFinString()) // Sorry, your credit is running low
	}

	// Send pre query payment information
	//fmt.Println(addrobj, amount)
	msg := &protocol.MsgRequestPrequeryPayment{
		PayAmount:        *amount,
		PaySatoshi:       satoshi,
		PayeeChannelAddr: fields.CreateStringMax255(addr),
	}
	err := protocol.SendMsg(c.user.servicerStreamSide.ChannelSide.WsConn, msg)
	if err != nil {
		return "SendMsg Error: " + err.Error()
	}
	c.statusMutex.Lock()
	c.pendingPaymentObj = &pendingPayment{
		address:     *acc,
		amount:      *amount,
		satoshi:     fields.NewSatoshiVariation(sts),
		prequeryMsg: nil,
	}
	c.statusMutex.Unlock()

	// no error
	return ""
}

// Close automatic collection
func (c *ChannelPayClient) BindFuncChangeAutoCollection(isopen int) {
	if isopen == 0 {
		// Close collection
		c.user.servicerStreamSide.ChannelSide.StartCloseAutoCollectionStatus() // Enable status
	} else {
		// Open collection
		c.user.servicerStreamSide.ChannelSide.ClearCloseAutoCollectionStatus() // Clear mark
	}
}

// Show payment errors
func (c *ChannelPayClient) ShowPaymentErrorString(err string) {
	c.payui.Eval(fmt.Sprintf(`ShowPaymentError("%s")`, strings.Replace(err, `"`, ``, -1)))
}

// Show log
func (c *ChannelPayClient) ShowLogString(log string, isok bool, iserr bool) {
	okmark := "0"
	if isok {
		okmark = "1"
	}
	errmark := "0"
	if iserr {
		errmark = "1"
	}
	c.payui.Eval(fmt.Sprintf(`ShowLogOnPrint("%s", %s, %s)`, strings.Replace(log, `"`, ``, -1), okmark, errmark))
}

func (c *ChannelPayClient) ShowStatusLog(log string) {
	c.payui.Eval(fmt.Sprintf(`ShowStatusLog("%s")`, strings.Replace(log, `"`, ``, -1)))
}

// Display interface
func (c *ChannelPayClient) ShowWindow() error {

	// Create UI with basic HTML passed via data URI
	sysType := runtime.GOOS
	ww := 965
	wh := 665
	if sysType == "windows" {
		ww = 995 // The win system avoids scrollbars
		wh = 674
	}
	ui, err := lorca.New("", "", ww, wh)
	if err != nil {
		log.Fatal(err)
	}
	err = ui.Load("data:text/html," + url.PathEscape(AccUIhtmlContent))
	if err != nil {
		log.Fatal(err)
	}

	c.payui = ui // ui

	// Binding operation function

	// Switch automatic collection
	ui.Bind("ChangeAutoCollection", c.BindFuncChangeAutoCollection)
	// Pre query payment
	ui.Bind("PrequeryPayment", c.BindFuncPrequeryPayment)
	// Confirm or cancel start payment
	ui.Bind("ConfirmPayment", c.BindFuncConfirmPayment)
	ui.Bind("CancelPayment", c.BindFuncCancelPayment)

	// Initialize account information
	ui.Eval(fmt.Sprintf(`InitAccount("%s","%s")`,
		c.user.selfAddr.ChannelId.ToHex(), c.user.selfAddr.ToReadable(false)))
	//fmt.Println(ui.Eval("2+2").Int())

	// Update balance
	c.UpdateBalanceShow()

	go func() {
		<-ui.Done()
		//fmt.Println("!!!!!!!!!!!!!!!!!!")
		// sign out
		c.user.Logout()
		if c.loginWindow != nil {
			c.loginWindow.Show() // Redisplay login window
		}
	}()

	return nil
}

/* 显示界面
func (c *ChannelPayClient) ShowWindow_old() error {

	if c.user == nil {
		panic("user   *ChannelPayUser == nil")
	}

	// Show login window
	objsleft := container.NewVBox()
	objsright := container.NewVBox()
	sizevb := fyne.Size{
		Width:  400,
		Height: 660,
	}



	balance := canvas.NewText("Balance: ", theme.TextColor())
	balance.Color = theme.PrimaryColorNamed("green")
	balance.TextSize = 18
	objsleft.Add(balance)




	objsright.Add(widget.NewLabel("123"))

	// Left and right layout
	objsleft.Resize(sizevb)
	objsleft.Move(fyne.Position{40,40})
	objsright.Resize(sizevb)
	objsright.Move(fyne.Position{40,40})
	wraply := fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		objsleft, objsright)


	wraply.Refresh()

	// Window layout
	// size
	wsize := &fyne.Size{
		Width:  960,
		Height: 740,
	}

	// Create and display windows
	c.window = NewVScrollWindowAndShow(c.app, wsize, wraply, "Channel pay and collect")
	c.window.SetPadded(true)

	// Intercept shutdown events
	c.window.SetCloseIntercept(func() {
		if c.user == nil || c.user.IsClosed() {
			c.window.Close() // Close without login
			return
		}
		// Ask whether to close
		dia := dialog.NewConfirm("Attention", "You can't collect after closing. Do you want to close the window and logout?", func(b bool) {
			if b {
				c.window.Close() // Confirm close
			}
		}, c.window)
		dia.Show() // display

	})

	return nil
}
*/
