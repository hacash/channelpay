package client

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/widget"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/account"
	"sync"
)

func CreateShowRunLoginWindow(app fyne.App) fyne.Window {

	windowSize := fyne.Size{
		Width:  460,
		Height: 640,
	}

	// 显示登录窗口
	objs := container.NewVBox()

	// title
	loginTitle := widget.NewLabel("\nChannel Account Login\n")
	loginTitle.Alignment = fyne.TextAlignCenter // 居中对齐
	loginTitle.TextStyle = fyne.TextStyle{
		Bold: true,
	}
	objs.Add(loginTitle)

	// label: address
	objs.Add(widget.NewLabel("Hacash Channel Address:"))
	inputAddr := widget.NewEntry()
	inputAddr.MultiLine = true
	inputAddr.Wrapping = fyne.TextWrapBreak
	inputAddr.SetPlaceHolder("example: 1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9_6814e443c4fe0615d3ea13732b313259_HACorg")
	inputAddr.Cursor()
	objs.Add(inputAddr)

	// 密码or私钥
	objs.Add(widget.NewLabel("\nPrivate key or Password:"))
	inputPrikey := widget.NewPasswordEntry()
	objs.Add(inputPrikey)

	objs.Add(widget.NewLabel("\n"))
	// login btn
	loginBtn := widget.NewButton("Login", nil)
	objs.Add(loginBtn)

	// error show
	errorshow := widget.NewEntry()
	errorshow.MultiLine = true
	errorshow.Wrapping = fyne.TextWrapBreak
	objs.Add(errorshow)

	// 显示窗口
	window := NewVScrollWindowAndShow(app, &windowSize, objs, "Hacash Channel Chain Payment User Login")

	// 点击登录按钮
	loginBtn.OnTapped = func() {
		// 执行登录
		e := HandlerLogin(inputAddr.Text, inputPrikey.Text, app, window)
		if e != nil {
			errorshow.SetText(e.Error())
		} else {
			// 登录成功，清空数据
			if DevDebug == false {
				inputAddr.SetText("")
				inputPrikey.SetText("")
			}
			errorshow.SetText("")
		}
		errorshow.Refresh()
	}

	return window
}

// 执行登录
func HandlerLogin(addr, prikeyorpassword string, app fyne.App, window fyne.Window) error {

	// 必填
	if len(addr) == 0 {
		return fmt.Errorf("Please enter the channel address.")
	}
	if len(prikeyorpassword) == 0 {
		return fmt.Errorf("Please enter the Private key or Password.")
	}

	// 检查地址格式
	addrobj, e := protocol.ParseChannelAccountAddress(addr)
	if e != nil {
		return fmt.Errorf("Parse channel account address error: \n%s", e.Error())
	}

	// 检查私钥
	loginAcc := account.GetAccountByPrivateKeyOrPassword(prikeyorpassword)
	if addrobj.Address.NotEqual(loginAcc.Address) {
		return fmt.Errorf("Private key or Password error.")
	}

	// 向网络请求服务商ip地址解析和通道状态数据
	apiurl := "https://hcpu.hacash.org"
	if DevDebug {
		apiurl = "http://127.0.0.1:3355"
	}
	chaninfo, nodeinfo, e := protocol.RequestChannelAndSernodeInfoFromLoginResolutionApi(
		apiurl, addrobj.ChannelId, addrobj.ServicerName.Value())
	if e != nil {
		return fmt.Errorf("request %s login resolution error: %s", apiurl, e.Error())
	}

	// 请求完毕
	//fmt.Println(chaninfo.Status, nodeinfo.Gateway.Value())

	// 创建用户端
	userObj := CreateChannelPayUser(loginAcc, addrobj, chaninfo)

	dtitle := "Reconciliation bill check"
	// 等待确认对话框
	waitConfirmDialog := func(msgcon string, cbfyes, cbfno func()) bool {
		var resck bool
		next := sync.WaitGroup{}
		next.Add(1)
		dia := dialog.NewConfirm(dtitle, msgcon, func(b bool) {
			resck = b
			if b && cbfyes != nil {
				cbfyes() // yes 回调
			}
			if !b && cbfno != nil {
				cbfno() // no 回调
			}
			next.Done()
		}, window)
		dia.Show()  // show 才有事件响应
		next.Wait() // 等待
		return resck
	}

	// 读取本地票据
	localbill, e := userObj.LoadLastBillFromDisk()

	// 检查通道已经重启，票据作废
	if localbill != nil {
		l1 := localbill.GetReuseVersion()
		if l1 != uint32(chaninfo.ReuseVersion) {
			// 本地不存在票据，是否使用远程对账票据
			waitConfirmDialog("Your local reconciliation bill has expired. Are you sure to delete it?", func() {
				// 确认删除本地过期票据
				userObj.DeleteLastBillOnDisk() // 删除
			}, nil)
		}
	}

	// 开始登录流程
	wsptcl := "wss"
	if DevDebug {
		wsptcl = "ws"
	}
	wsurl := fmt.Sprintf("%s://%s/customer/connect", wsptcl, nodeinfo.Gateway.Value())
	e = userObj.ConnectServicer(wsurl)
	if e != nil {
		// 登录失败
		return fmt.Errorf("Connect servicer %s error: %s", wsurl, e.Error())
	}
	remotebill := userObj.GetReconciliationBalanceBillAfterLoginFromRemote()

	// 票据对比
	if remotebill != nil {
		// 远程存在票据
		var msgcon string = ""
		if localbill == nil {
			// 本地不存在对账票据，是否使用你的支付服务商发送的对账票据？
			msgcon = "There is no reconciliation bill in the local area. Do you want to use the reconciliation bill sent by your payment service provider?"
		} else {
			// 本地存在票据，对比票据版本
			l1, l2 := localbill.GetReuseVersionAndAutoNumber()
			r1, r2 := remotebill.GetReuseVersionAndAutoNumber()
			if l1 < r1 || l2 < r2 {
				// 本地的对账票据序列号落后于远程，是否强制使用你的支付服务商发送的对账票据？
				msgcon = "The serial number of the local reconciliation bill lags behind that of the remote reconciliation bill. Is it mandatory to use the reconciliation bill sent by your payment service provider?"
			} else if l1 > r1 || l2 > r2 {
				// 本地的对账票据序列号高于远程，是否强制使用你的支付服务商发送的对账票据？
				msgcon1 := "The serial number of the local reconciliation bill is higher than that of the remote reconciliation bill. Is it mandatory to use the reconciliation bill sent by your payment service provider?"
				useok := waitConfirmDialog(msgcon1, func() {
					userObj.SaveLastBillToDisk(remotebill) // 将远程票据保存到本地
				}, nil)
				if !useok {
					userObj.Logout()
					return nil // 不使用，则退出
				}
			}
		}
		if len(msgcon) > 0 {
			// 本地不存在票据或者票据落后，是否使用远程对账票据
			useok := waitConfirmDialog(msgcon, func() {
				userObj.SaveLastBillToDisk(remotebill) // 将远程票据保存到本地
			}, nil)
			if !useok {
				userObj.Logout()
				return nil // 不使用，则退出
			}
		} else {
			// 序列号检查一致，继续通过
		}

	} else {
		// 远程不存在票据
		if localbill != nil {
			// 本地存在票据
			waitConfirmDialog("The bill does not exist remotely, but exists locally. Please contact your payment service provider for processing.", nil, nil)
			userObj.Logout()
			return nil // 不使用，则退出
		} else {
			// 本地也不存在票据，通过
		}
	}

	// 打开通道支付界面
	client := CreateChannelPayClient(app, userObj, window)
	e = client.ShowWindow()
	if e != nil {
		userObj.Logout() // 显示发生错误，退出
		return e
	}

	// 打开消息监听
	client.startMsgHandler()

	// 开始监听消息
	userObj.upstreamSide.ChannelSide.StartMessageListen()

	// 登录窗口影藏
	if DevDebug == false {
		window.Hide()
	}

	// 成功完成
	return nil
}
