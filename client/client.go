package client

import (
	"encoding/hex"
	"fmt"
	"fyne.io/fyne"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/zserge/lorca"
	"log"
	"net/url"
	"strings"
	"sync"
)

// 待支付结构
type pendingPayment struct {
	address     protocol.ChannelAccountAddress
	amount      fields.Amount
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
	user *ChannelPayUser // 用户端
	// 待支付缓存数据
	pendingPaymentObj *pendingPayment

	// 状态锁定
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

// 更新界面显示
func (c *ChannelPayClient) UpdateBalanceShow() {
	cside := c.user.upstreamSide.ChannelSide
	bill := cside.GetReconciliationBill()
	billhex := ""
	if bill != nil {
		bts, _ := bill.SerializeWithTypeCode()
		billhex = hex.EncodeToString(bts)
	}
	c.payui.Eval(fmt.Sprintf(`UpdateBalance("%s","%s",%d,%d,"%s")`,
		cside.GetChannelCapacityAmountOfOur().ToFinString(),
		cside.GetChannelCapacityAmountOfRemote().ToFinString(),
		cside.GetAvailableReuseVersion(),
		cside.GetAvailableAutoNumber(),
		billhex,
	))
}

// 启动交易
func (c *ChannelPayClient) BindFuncPrequeryPayment(addr, amt string) string {
	//fmt.Println("BindFuncInitiatePayment:", addr, amt)
	if len(addr) == 0 {
		return "Please enter the address."
	}
	if len(amt) == 0 {
		return "Please enter the amount."
	}
	acc, e := protocol.ParseChannelAccountAddress(addr)
	if e != nil {
		return fmt.Sprintf("Address format error: %s", e.Error()) // 错误
	}
	amount, e := fields.NewAmountFromStringUnsafe(amt)
	if e != nil {
		return fmt.Sprintf("Amount format error: %s", e.Error()) // 错误
	}
	// 余额检查
	amtcap := c.user.upstreamSide.ChannelSide.GetChannelCapacityAmountOfOur()
	if amtcap.LessThan(amount) {
		return fmt.Sprintf("Balance %s not enough for transfer %s",
			amtcap.ToFinString(), amount.ToFinString()) // 余额不足
	}
	// 发送预查询支付信息
	//fmt.Println(addrobj, amount)
	msg := &protocol.MsgRequestPrequeryPayment{
		PayAmount:        *amount,
		PayeeChannelAddr: fields.CreateStringMax255(addr),
	}
	err := protocol.SendMsg(c.user.upstreamSide.ChannelSide.WsConn, msg)
	if err != nil {
		return "SendMsg Error: " + err.Error()
	}
	c.statusMutex.Lock()
	c.pendingPaymentObj = &pendingPayment{
		address:     *acc,
		amount:      *amount,
		prequeryMsg: nil,
	}
	c.statusMutex.Unlock()

	// no error
	return ""
}

// 是否关闭自动收款
func (c *ChannelPayClient) BindFuncChangeAutoCollection(isopen int) {
	if isopen == 0 {
		// 关闭收款
		c.user.upstreamSide.ChannelSide.StartCloseAutoCollectionStatus() // 启用状态
	} else {
		// 开启收款
		c.user.upstreamSide.ChannelSide.ClearCloseAutoCollectionStatus() // 清除标记
	}
}

// 显示支付错误
func (c *ChannelPayClient) ShowPaymentErrorString(err string) {
	c.payui.Eval(fmt.Sprintf(`ShowPaymentError("%s")`, strings.Replace(err, `"`, ``, -1)))
}

// 显示日志
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

// 显示界面
func (c *ChannelPayClient) ShowWindow() error {

	// Create UI with basic HTML passed via data URI
	ui, err := lorca.New("", "", 962, 642)
	if err != nil {
		log.Fatal(err)
	}
	err = ui.Load("data:text/html," + url.PathEscape(AccUIhtmlContent))
	if err != nil {
		log.Fatal(err)
	}

	c.payui = ui // ui

	// 绑定操作函数

	// 开关自动收款
	ui.Bind("ChangeAutoCollection", c.BindFuncChangeAutoCollection)
	// 预查询支付
	ui.Bind("PrequeryPayment", c.BindFuncPrequeryPayment)
	// 确认或取消启动支付
	ui.Bind("ConfirmPayment", c.BindFuncConfirmPayment)
	ui.Bind("CancelPayment", c.BindFuncCancelPayment)

	// 初始化账户信息
	ui.Eval(fmt.Sprintf(`InitAccount("%s","%s")`,
		c.user.selfAddr.ChannelId.ToHex(), c.user.selfAddr.ToReadable(false)))
	//fmt.Println(ui.Eval("2+2").Int())

	// 更新余额
	c.UpdateBalanceShow()

	go func() {
		<-ui.Done()
		//fmt.Println("!!!!!!!!!!!!!!!!!!")
		// 退出
		c.user.Logout()
		if c.loginWindow != nil {
			c.loginWindow.Show() // 重新显示登录窗口
		}
	}()

	return nil
}

/* 显示界面
func (c *ChannelPayClient) ShowWindow_old() error {

	if c.user == nil {
		panic("user   *ChannelPayUser == nil")
	}

	// 显示登录窗口
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

	// 左右布局
	objsleft.Resize(sizevb)
	objsleft.Move(fyne.Position{40,40})
	objsright.Resize(sizevb)
	objsright.Move(fyne.Position{40,40})
	wraply := fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		objsleft, objsright)


	wraply.Refresh()

	// 窗口布局
	// 尺寸
	wsize := &fyne.Size{
		Width:  960,
		Height: 740,
	}

	// 创建并显示窗口
	c.window = NewVScrollWindowAndShow(c.app, wsize, wraply, "Channel pay and collect")
	c.window.SetPadded(true)

	// 拦截关闭事件
	c.window.SetCloseIntercept(func() {
		if c.user == nil || c.user.IsClosed() {
			c.window.Close() // 未登录直接关闭
			return
		}
		// 询问是否关闭
		dia := dialog.NewConfirm("Attention", "You can't collect after closing. Do you want to close the window and logout?", func(b bool) {
			if b {
				c.window.Close() // 确认关闭
			}
		}, c.window)
		dia.Show() // 显示

	})

	return nil
}
*/
