package client

import (
	"encoding/hex"
	"fmt"
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/account"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/stores"
	"strings"
	"sync"
	"time"
)

const (
	windowWidth = 500
)

func CreateShowRunLoginWindow(app fyne.App) fyne.Window {

	windowSize := fyne.Size{
		Width:  windowWidth,
		Height: 700,
	}

	window := app.NewWindow("Hacash Channel Chain Payment User Login")

	// 显示登录窗口
	loginBox := createLoginTab(app, window)
	billBox := createOutputBillTab(app, window)

	tabs := container.NewAppTabs(
		container.NewTabItem("Login", loginBox),
		container.NewTabItem("Export bill", billBox),
	)

	// 显示窗口
	NewVScrollAndShowWindow(window, &windowSize, tabs)
	return window

}

// 创建导出票据窗口
func createOutputBillTab(app fyne.App, window fyne.Window) *fyne.Container {

	objs := container.NewVBox()

	// title
	title := widget.NewLabel("\nExport locally stored payment bill")
	title.Alignment = fyne.TextAlignCenter // 居中对齐
	title.TextStyle = fyne.TextStyle{
		Bold: true,
	}
	objs.Add(title)

	// 通道id
	objs.Add(widget.NewLabel("Channel ID:"))
	inputChannelID := widget.NewEntry()
	objs.Add(inputChannelID)

	// export btn
	okBtn := widget.NewButton("Export", nil)
	objs.Add(okBtn)

	// results show
	resshow := widget.NewMultiLineEntry()
	resshow.Wrapping = fyne.TextWrapBreak
	objs.Add(resshow)

	// 点击导出按钮
	okBtn.OnTapped = func() {
		// 查询本地是否存在票据
		chanid, e := hex.DecodeString(inputChannelID.Text)
		if e != nil || chanid == nil || len(chanid) != stores.ChannelIdLength {
			resshow.SetText("Channel ID error.")
			resshow.Refresh()
			return
		}
		bill, e := LoadBillFromDisk(chanid)
		if e != nil || bill == nil {
			resshow.SetText("Not find this channel.")
			resshow.Refresh()
			return
		}
		// 显示票据
		content := "Bill exported successfully!\n"
		content += "---- bill hex data start ----\n"
		bldt, _ := bill.SerializeWithTypeCode()
		content += hex.EncodeToString(bldt)
		content += "\n---- bill hex data end ----"
		resshow.SetText(content)
		resshow.Refresh()
	}

	return objs
}

// 创建登录窗口
func createLoginTab(app fyne.App, window fyne.Window) *fyne.Container {

	objs := container.NewVBox()

	// title
	loginTitle := widget.NewLabel("\nChannel account login")
	loginTitle.Alignment = fyne.TextAlignCenter // 居中对齐
	loginTitle.TextStyle = fyne.TextStyle{
		Bold: true,
	}
	objs.Add(loginTitle)

	// label: address
	objs.Add(widget.NewLabel("Hacash Channel Address:"))
	inputAddr := widget.NewMultiLineEntry()
	inputAddr.MultiLine = true
	inputAddr.Wrapping = fyne.TextWrapBreak
	inputAddr.SetPlaceHolder("example: 1MzNY1oA3kfgYi75zquj3SRUPYztzXHzK9_6814e443c4fe0615d3ea13732b313259_HACorg")
	inputAddr.Cursor()
	objs.Add(inputAddr)

	// 密码or私钥
	objs.Add(widget.NewLabel("\nPrivate key or Password:"))
	inputPrikey := widget.NewPasswordEntry()
	objs.Add(inputPrikey)

	// 对账票据
	objs.Add(widget.NewLabel("\nReconciliation or payment bill:"))
	inputBill := widget.NewMultiLineEntry()
	inputBill.Wrapping = fyne.TextWrapBreak
	inputBill.Scroll = true
	inputBill.SetPlaceHolder("Optional: reconciliation or payment bill hex data")
	inputBill.Refresh()
	objs.Add(inputBill)

	//objs.Add(widget.NewLabel("\n"))
	// login btn
	loginBtn := widget.NewButton("Login", nil)
	objs.Add(loginBtn)

	// error show
	errorshow := widget.NewEntry()
	errorshow.MultiLine = true
	errorshow.Wrapping = fyne.TextWrapBreak
	objs.Add(errorshow)

	// 点击登录按钮
	loginBtn.OnTapped = func() {
		go func() {
			// 执行登录
			e := HandlerLogin(inputAddr.Text, inputPrikey.Text, inputBill.Text, app, window, loginBtn)
			if e != nil {
				errorshow.SetText(e.Error())
			} else {
				// 登录成功，清空数据
				if DevDebug == false {
					//inputAddr.SetText("")
					inputPrikey.SetText("")
					inputBill.SetText("")
				}
				errorshow.SetText("")
			}
			errorshow.Refresh()
		}()
	}

	// 返回
	return objs
}

func trimInput(addr string) string {
	return strings.Trim(addr, " \n")
}

// 执行登录
var isInHandlerLoginState = false

func HandlerLogin(addr, prikeyorpassword, billhex string, app fyne.App, window fyne.Window, loginBtn *widget.Button) error {
	if isInHandlerLoginState {
		return nil // 正在处理中
	}
	if loginBtn != nil {
		loginBtn.SetText("waiting...")
	}
	isInHandlerLoginState = true
	go func() {
		time.Sleep(time.Second * 3)
		if loginBtn != nil {
			loginBtn.SetText("Login")
		}
		isInHandlerLoginState = false
	}()

	// 地址和私钥去掉前后空格和换行
	addr = trimInput(addr)
	prikeyorpassword = trimInput(prikeyorpassword)
	billhex = trimInput(billhex)

	// 必填
	if len(addr) == 0 {
		return fmt.Errorf("Please enter the channel address.")
	}
	if len(prikeyorpassword) == 0 {
		return fmt.Errorf("Please enter the Private key or Password.")
	}

	// 选填的对账票据
	var inputbillobj channel.ReconciliationBalanceBill = nil
	if len(billhex) > 0 {
		billdata, e := hex.DecodeString(billhex)
		if e != nil {
			return fmt.Errorf("billdata format error")
		}
		inputbillobj, _, e = channel.ParseReconciliationBalanceBillByPrefixTypeCode(billdata, 0)
		if e != nil {
			return fmt.Errorf("billdata parse error: %s", e.Error())
		}
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
	apiurl := GetLoginResolutionApiDomain()
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
		if window == nil {
			cbfyes()
			return true // 测试是直接选 yes
		}
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
		dia.Resize(fyne.Size{
			Width:  windowWidth - 20,
			Height: 200,
		})
		dia.Show()
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
			waitConfirmDialog("Your local reconciliation bill has expired.\n Are you sure to delete it?", func() {
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

	// 输入票据对比
	if inputbillobj != nil {
		if localbill == nil {
			// 本地不存在票据，直接使用输入的票据
			localbill = inputbillobj
		} else {
			// 对比票据版本
			if inputbillobj.GetReuseVersion() < localbill.GetReuseVersion() ||
				inputbillobj.GetAutoNumber() < localbill.GetAutoNumber() {
				// 输入的票据版本过小询问是否强制使用
				waitConfirmDialog("The reconciliation bill version entered\n is older than that saved locally. \nDo you want to force the entered\n bill to be used?", func() {
					localbill = inputbillobj
				}, nil)
			} else {
				// 输入的票据版本大于本地保存，不询问，直接使用
				localbill = inputbillobj
			}
		}
	}

	// 远程票据对比
	if remotebill != nil {
		// 远程存在票据
		var msgcon string = ""
		if localbill == nil {
			// 本地不存在对账票据，是否使用你的支付服务商发送的对账票据？
			msgcon = "There is no reconciliation bill \nin the local area. \nDo you want to use the reconciliation bill \nsent by your payment service provider?"
		} else {
			// 本地存在票据，对比票据版本
			l1, l2 := localbill.GetReuseVersionAndAutoNumber()
			r1, r2 := remotebill.GetReuseVersionAndAutoNumber()
			if l1 < r1 || l2 < r2 {
				// 本地的对账票据序列号落后于远程，是否强制使用你的支付服务商发送的对账票据？
				msgcon = "The serial number of the local \nreconciliation bill lags behind that \nof the remote reconciliation bill. \nIs it mandatory to use the reconciliation bill\n sent by your payment service provider?"
			} else if l1 > r1 || l2 > r2 {
				// 本地的对账票据序列号高于远程，是否强制使用你的支付服务商发送的对账票据？
				msgcon = "The serial number of the local reconciliation bill\n is higher than that of the remote reconciliation bill.\n Is it mandatory to use the reconciliation bill\n sent by your payment service provider?"
			}
		}
		if len(msgcon) > 0 {
			// 本地不存在票据或者票据落后，是否使用远程对账票据
			useok := waitConfirmDialog(msgcon, func() {
				userObj.SaveLastBillToDisk(remotebill) // 将远程票据保存到本地
				localbill = remotebill                 // 使用远程票据
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
			waitConfirmDialog("The bill does not exist remotely,\n but exists locally.\n Please contact your payment\n service provider for processing.", nil, nil)
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

	// 设置票据
	userObj.servicerStreamSide.ChannelSide.SetReconciliationBill(localbill)

	// 打开消息监听
	client.startMsgHandler()

	// 开始监听消息
	userObj.servicerStreamSide.ChannelSide.StartMessageListen()

	// 登录窗口影藏
	if DevDebug == false {
		window.Hide()
	}

	// 成功完成
	return nil
}
