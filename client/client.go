package client

import (
	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/widget"
)

/**
 * 通道链支付客户端
 */

type ChannelPayClient struct {
	app    fyne.App
	window fyne.Window
	user   *ChannelPayUser // 用户端

}

func CreateChannelPayClient(app fyne.App, user *ChannelPayUser) *ChannelPayClient {
	return &ChannelPayClient{
		app:  app,
		user: user,
	}
}

// 显示界面
func (c *ChannelPayClient) ShowWindow() error {

	// 显示登录窗口
	objs := container.NewVBox()

	// 窗口布局
	objs.Add(widget.NewLabel("balance: ㄜ1:248"))

	wsize := &fyne.Size{
		Width:  800,
		Height: 600,
	}

	// 创建并显示窗口
	c.window = NewVScrollWindowAndShow(c.app, wsize, objs, "Channel pay and collect")

	// 拦截关闭事件
	c.window.SetCloseIntercept(func() {
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
