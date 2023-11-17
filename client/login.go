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

	// Show login window
	loginBox := createLoginTab(app, window)
	billBox := createOutputBillTab(app, window)

	tabs := container.NewAppTabs(
		container.NewTabItem("Login", loginBox),
		container.NewTabItem("Export bill", billBox),
	)

	// Display window
	NewVScrollAndShowWindow(window, &windowSize, tabs)
	return window

}

// Create export ticket window
func createOutputBillTab(app fyne.App, window fyne.Window) *fyne.Container {

	objs := container.NewVBox()

	// title
	title := widget.NewLabel("\nExport locally stored payment bill")
	title.Alignment = fyne.TextAlignCenter // Center align
	title.TextStyle = fyne.TextStyle{
		Bold: true,
	}
	objs.Add(title)

	// Channel ID
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

	// Click the Export button
	okBtn.OnTapped = func() {
		// Query whether bills exist locally
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
		// Show tickets
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

// Create login window
func createLoginTab(app fyne.App, window fyne.Window) *fyne.Container {

	objs := container.NewVBox()

	// title
	loginTitle := widget.NewLabel("\nChannel account login")
	loginTitle.Alignment = fyne.TextAlignCenter // Center align
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

	// Password or private key
	objs.Add(widget.NewLabel("Private key or Password:"))
	inputPrikey := widget.NewPasswordEntry()
	objs.Add(inputPrikey)

	// Reconciliation bill
	objs.Add(widget.NewLabel("Reconciliation or payment bill:"))
	inputBill := widget.NewMultiLineEntry()
	inputBill.Wrapping = fyne.TextWrapBreak
	inputBill.Scroll = true
	inputBill.SetPlaceHolder("Optional: reconciliation or payment bill hex data")
	inputBill.Refresh()
	objs.Add(inputBill)

	// route api url
	objs.Add(widget.NewLabel("Route Server URL:"))
	inputRouteURL := widget.NewEntry()
	inputRouteURL.SetPlaceHolder("Optional: can use test url")
	objs.Add(inputRouteURL)

	objs.Add(widget.NewLabel("\n"))
	// login btn
	loginBtn := widget.NewButton("Login", nil)
	objs.Add(loginBtn)

	// error show
	errorshow := widget.NewEntry()
	errorshow.MultiLine = true
	errorshow.Wrapping = fyne.TextWrapBreak
	objs.Add(errorshow)

	// Click the login button
	loginBtn.OnTapped = func() {
		go func() {
			// Execute login
			e := HandlerLogin(inputAddr.Text, inputPrikey.Text, inputBill.Text, inputRouteURL.Text, app, window, loginBtn)
			if e != nil {
				errorshow.SetText(e.Error())
			} else {
				// Login successful, clear data
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

	// return
	return objs
}

func trimInput(addr string) string {
	return strings.Trim(addr, " \n")
}

// Execute login
var isInHandlerLoginState = false

func HandlerLogin(addr, prikeyorpassword, billhex, routeurl string, app fyne.App, window fyne.Window, loginBtn *widget.Button) error {
	if isInHandlerLoginState {
		return nil // Processing
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

	// Address and private key remove space before and after and line feed
	addr = trimInput(addr)
	prikeyorpassword = trimInput(prikeyorpassword)
	billhex = trimInput(billhex)
	routeurl = trimInput(routeurl)

	// Required
	if len(addr) == 0 {
		return fmt.Errorf("Please enter the channel address.")
	}
	if len(prikeyorpassword) == 0 {
		return fmt.Errorf("Please enter the Private key or Password.")
	}

	// Optional reconciliation bill
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

	// Check address format
	addrobj, e := protocol.ParseChannelAccountAddress(addr)
	if e != nil {
		return fmt.Errorf("Parse channel account address error: \n%s", e.Error())
	}

	// Check private key
	loginAcc := account.GetAccountByPrivateKeyOrPassword(prikeyorpassword)
	if addrobj.Address.NotEqual(loginAcc.Address) {
		return fmt.Errorf("Private key or Password error.")
	}

	// Requesting IP address resolution and channel status data from the network service provider
	if len(routeurl) > 0 {
		SetLoginResolutionApiDomain(routeurl) // change url
	}
	apiurl := GetLoginResolutionApiDomain()
	chaninfo, nodeinfo, e := protocol.RequestChannelAndSernodeInfoFromLoginResolutionApi(
		apiurl, addrobj.ChannelId, addrobj.ServicerName.Value())
	if e != nil {
		return fmt.Errorf("request %s login resolution error: %s", apiurl, e.Error())
	}

	// Request complete
	//fmt.Println(chaninfo.Status, nodeinfo.Gateway.Value())

	// Create client
	userObj := CreateChannelPayUser(loginAcc, addrobj, chaninfo)

	dtitle := "Reconciliation bill check"
	// Waiting for confirmation dialog
	waitConfirmDialog := func(msgcon string, cbfyes, cbfno func()) bool {
		if window == nil {
			cbfyes()
			return true // The test is to directly select Yes
		}
		var resck bool
		next := sync.WaitGroup{}
		next.Add(1)
		dia := dialog.NewConfirm(dtitle, msgcon, func(b bool) {
			resck = b
			if b && cbfyes != nil {
				cbfyes() // Yes callback
			}
			if !b && cbfno != nil {
				cbfno() // No callback
			}
			next.Done()
		}, window)
		dia.Resize(fyne.Size{
			Width:  windowWidth - 20,
			Height: 200,
		})
		dia.Show()
		next.Wait() // wait for
		return resck
	}

	// Read local ticket
	localbill, e := userObj.LoadLastBillFromDisk()

	// Check that the channel has been restarted and the bill is voided
	if localbill != nil {
		l1 := localbill.GetReuseVersion()
		if l1 != uint32(chaninfo.ReuseVersion) {
			// Bill does not exist locally, whether to use remote reconciliation bill
			waitConfirmDialog("Your local reconciliation bill has expired.\n Are you sure to delete it?", func() {
				// Confirm to delete local overdue bills
				userObj.DeleteLastBillOnDisk() // delete
			}, nil)
		}
	}

	// Start login process
	wsptcl := "wss"
	if DevDebug || strings.HasPrefix(apiurl, "http://") {
		wsptcl = "ws"
	}
	wsurl := fmt.Sprintf("%s://%s/customer/connect", wsptcl, nodeinfo.Gateway.Value())
	e = userObj.ConnectServicer(wsurl)
	if e != nil {
		// Login failed
		return fmt.Errorf("Connect servicer %s error: %s", wsurl, e.Error())
	}
	remotebill := userObj.GetReconciliationBalanceBillAfterLoginFromRemote()

	// Input bill comparison
	if inputbillobj != nil {
		if localbill == nil {
			// There is no bill in the local area, and the entered bill is used directly
			localbill = inputbillobj
		} else {
			// Compare bill versions
			if inputbillobj.GetReuseVersion() < localbill.GetReuseVersion() ||
				inputbillobj.GetAutoNumber() < localbill.GetAutoNumber() {
				// The entered bill version is too small. Ask whether to force use
				waitConfirmDialog("The reconciliation bill version entered\n is older than that saved locally. \nDo you want to force the entered\n bill to be used?", func() {
					localbill = inputbillobj
				}, nil)
			} else {
				// The entered bill version is greater than that saved locally. You can use it directly without asking
				localbill = inputbillobj
			}
		}
	}

	// Remote bill comparison
	if remotebill != nil {
		// Remote presence ticket
		var msgcon string = ""
		if localbill == nil {
			// There is no reconciliation bill in the local area. Do you want to use the reconciliation bill sent by your payment service provider?
			msgcon = "There is no reconciliation bill \nin the local area. \nDo you want to use the reconciliation bill \nsent by your payment service provider?"
		} else {
			// Bills exist locally. Compare bill versions
			l1, l2 := localbill.GetReuseVersionAndAutoNumber()
			r1, r2 := remotebill.GetReuseVersionAndAutoNumber()
			if l1 < r1 || l2 < r2 {
				// The serial number of the local reconciliation bill lags behind that of the remote one. Is it mandatory to use the reconciliation bill sent by your payment service provider?
				msgcon = "The serial number of the local \nreconciliation bill lags behind that \nof the remote reconciliation bill. \nIs it mandatory to use the reconciliation bill\n sent by your payment service provider?"
			} else if l1 > r1 || l2 > r2 {
				// The serial number of the local reconciliation bill is higher than that of the remote reconciliation bill. Is it mandatory to use the reconciliation bill sent by your payment service provider?
				msgcon = "The serial number of the local reconciliation bill\n is higher than that of the remote reconciliation bill.\n Is it mandatory to use the reconciliation bill\n sent by your payment service provider?"
			}
		}
		if len(msgcon) > 0 {
			// Whether to use remote reconciliation bill if there is no bill in the local area or the bill is backward
			useok := waitConfirmDialog(msgcon, func() {
				userObj.SaveLastBillToDisk(remotebill) // Save remote ticket locally
				localbill = remotebill                 // Use remote ticket
			}, nil)
			if !useok {
				userObj.Logout()
				return nil // Exit if not used
			}
		} else {
			// Serial number is consistent, continue to pass
		}

	} else {
		// Remote ticket does not exist
		if localbill != nil {
			// Locally existing bills
			waitConfirmDialog("The bill does not exist remotely,\n but exists locally.\n Please contact your payment\n service provider for processing.", nil, nil)
			userObj.Logout()
			return nil // Exit if not used
		} else {
			// There is no bill in the local area, and the
		}
	}

	// Open channel payment interface
	client := CreateChannelPayClient(app, userObj, window)
	e = client.ShowWindow()
	if e != nil {
		userObj.Logout() // Display error, exit
		return e
	}

	// Set up ticket
	userObj.servicerStreamSide.ChannelSide.SetReconciliationBill(localbill)

	// Open message listening
	client.startMsgHandler()

	// Start listening for messages
	userObj.servicerStreamSide.ChannelSide.StartMessageListen()

	// Login window shadow
	if DevDebug == false {
		window.Hide()
	}

	// Successfully completed
	return nil
}
