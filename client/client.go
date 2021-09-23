package client

import (
	"fyne.io/fyne"
	"github.com/zserge/lorca"
	"log"
	"net/url"
)

/**
 * 通道链支付客户端
 */

type ChannelPayClient struct {
	app         fyne.App
	loginWindow fyne.Window
	payui       lorca.UI
	//window fyne.Window
	user *ChannelPayUser // 用户端

}

func CreateChannelPayClient(app fyne.App, user *ChannelPayUser, lgwd fyne.Window) *ChannelPayClient {
	return &ChannelPayClient{
		app:         app,
		loginWindow: lgwd,
		user:        user,
	}
}

// 显示界面
func (c *ChannelPayClient) ShowWindow() error {

	// Create UI with basic HTML passed via data URI
	ui, err := lorca.New("data:text/html,"+url.PathEscape(AccUIhtmlContent),
		"", 960, 640)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(ui.Eval("2+2").Int())

	go func() {
		<-ui.Done()
		// 退出
		c.user.Logout()
		if c.loginWindow != nil {
			c.loginWindow.Show() // 重新显示登录窗口
		}
	}()

	c.payui = ui // ui

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
