package client

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/account"
	"github.com/hacash/core/channel"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/sys"
	"github.com/hacash/mint/event"
	"github.com/hacash/node/websocket"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

const (
	billDir string = "hacash_channel_pay_bill_data"
)

/**
 * 通道链支付用户端
 */
type ChannelPayUser struct {
	changeMux sync.Mutex

	// 账户地址
	selfAcc  *account.Account
	selfAddr *protocol.ChannelAccountAddress
	chanInfo *protocol.RpcDataChannelInfo

	// 本地对账票据
	localLatestReconciliationBalanceBill channel.ReconciliationBalanceBill

	// 等待对账单
	waitRealtimeReconciliation *channel.OffChainFormPaymentChannelRealtimeReconciliation

	// 通道链
	servicerStreamSide *chanpay.RelayPaySettleNoder

	// 消息订阅
	msgSubObj event.Subscription

	// 是否已经关闭
	isClosed bool
}

func CreateChannelPayUser(acc *account.Account, addr *protocol.ChannelAccountAddress, cinfo *protocol.RpcDataChannelInfo) *ChannelPayUser {
	return &ChannelPayUser{
		selfAcc:  acc,
		selfAddr: addr,
		chanInfo: cinfo,
		isClosed: false,
	}
}

// 检查收款通道是否被占用
func (c *ChannelPayUser) IsInBusinessExclusive() bool {
	return c.servicerStreamSide.IsInBusinessExclusive()
}

// 其中状态独占
func (c *ChannelPayUser) StartBusinessExclusive() bool {
	return c.servicerStreamSide.StartBusinessExclusive()
}

// 解除状态独占
func (c *ChannelPayUser) ClearBusinessExclusive() {
	c.servicerStreamSide.ClearBusinessExclusive()
}

// 开始发送心跳包
func (c *ChannelPayUser) StartHeartbeat() {
	for {
		// 15秒一个心跳包
		time.Sleep(time.Second * 14)
		c.changeMux.Lock()
		if c.isClosed {
			c.changeMux.Unlock()
			return // 结束
		}
		c.changeMux.Unlock()
		// 发送心跳包，忽略错误
		//fmt.Println("protocol.SendMsg(c.servicerStreamSide.ChannelSide.WsConn, &protocol.MsgHeartbeat{})")
		protocol.SendMsg(c.servicerStreamSide.ChannelSide.WsConn, &protocol.MsgHeartbeat{})
		// 判断服务器返回的心跳
		lastBeatTime := c.servicerStreamSide.ChannelSide.GetLastestHeartbeatTime()
		tnck := time.Now().Unix() - 60
		if lastBeatTime.Unix() < tnck {
			// 60s一分钟都没有收到服务器的心跳包，自动断开
			c.Logout() // 退出
		}
	}
}

// 退出
func (c *ChannelPayUser) IsClosed() bool {
	return c.isClosed // 是否已经退出、关闭
}

func (c *ChannelPayUser) Logout() {
	c.changeMux.Lock()
	defer c.changeMux.Unlock()
	if c.isClosed {
		return // 已经退出
	}
	c.isClosed = true
	if c.servicerStreamSide != nil {
		wsconn := c.servicerStreamSide.ChannelSide.WsConn
		// 发送退出消息
		protocol.SendMsg(wsconn, &protocol.MsgCustomerLogout{
			PostBack: fields.CreateStringMax255(""),
		})
		wsconn.Close()
	}
	// 关闭订阅
	if c.msgSubObj != nil {
		c.msgSubObj.Unsubscribe()
		c.msgSubObj = nil
	}
}

func billfilepath(channelID fields.ChannelId) string {
	datadir := sys.AbsDir(billDir)
	os.MkdirAll(datadir, 0777)
	fname := path.Join(datadir, fmt.Sprintf("bill_%s.dat", channelID.ToHex()))
	return fname
}

func LoadBillFromDisk(channelID fields.ChannelId) (channel.ReconciliationBalanceBill, error) {
	var bill channel.ReconciliationBalanceBill = nil
	// 读取本地目录
	fname := billfilepath(channelID)
	fbts, e := ioutil.ReadFile(fname)
	if e == nil || len(fbts) > 0 {
		// 解析文件
		bill, _, e = channel.ParseReconciliationBalanceBillByPrefixTypeCode(fbts, 0)
	}
	// 返回
	return bill, nil
}

// 删除本地票据
func (c *ChannelPayUser) DeleteLastBillOnDisk() error {
	// 读取本地目录
	fname := billfilepath(c.selfAddr.ChannelId)
	c.localLatestReconciliationBalanceBill = nil
	// 删除文件
	return os.Remove(fname)
}

// 读取本地的对账票据
func (c *ChannelPayUser) LoadLastBillFromDisk() (channel.ReconciliationBalanceBill, error) {
	var bill channel.ReconciliationBalanceBill = nil
	// 读取本地目录
	ldbill, e := LoadBillFromDisk(c.selfAddr.ChannelId)
	if e == nil {
		bill = ldbill
	}
	c.localLatestReconciliationBalanceBill = bill
	// 返回
	return bill, nil
}

// 保存对账票据
func (c *ChannelPayUser) SaveLastBillToDisk(bill channel.ReconciliationBalanceBill) error {
	if bill == nil {
		return fmt.Errorf("bill is nil")
	}
	// 本地目录
	fname := billfilepath(c.selfAddr.ChannelId)
	fbts, e := channel.SerializeReconciliationBalanceBillWithPrefixTypeCode(bill)
	if e == nil || len(fbts) > 0 {
		// 保存
		e = ioutil.WriteFile(fname, fbts, 0777)
		if e != nil {
			return e // 返回错误
		}
	}
	c.localLatestReconciliationBalanceBill = bill
	// ok
	return nil
}

// 登陆后从远程取得对账票据
func (c *ChannelPayUser) GetReconciliationBalanceBillAfterLoginFromRemote() channel.ReconciliationBalanceBill {
	if c.servicerStreamSide != nil {
		return c.servicerStreamSide.ChannelSide.LatestReconciliationBalanceBill
	}
	return nil // 不存在
}

// 连接到服务端
func (c *ChannelPayUser) ConnectServicer(wsurl string) error {
	// 拨号
	wsconn, e := websocket.Dial(wsurl, "", "http://127.0.0.1")
	if e != nil {
		return fmt.Errorf("Connect servicer %s error: %s", wsurl, e.Error())
	}

	// 发送身份
	resmsg, _, e := protocol.SendMsgForResponseTimeout(wsconn, &protocol.MsgLogin{
		ProtocolVersion: fields.VarUint2(protocol.LatestProtocolVersion),
		ChannelId:       c.selfAddr.ChannelId,
		CustomerAddress: c.selfAddr.Address,
		LanguageSet:     fields.CreateStringMax255("US"),
	}, 6)
	if e != nil {
		wsconn.Close()
		return fmt.Errorf("Send login msg error: %s", e.Error())
	}

	// 登录错误
	resmsgty := resmsg.Type()
	if resmsgty == protocol.MsgTypeError {
		wsconn.Close()
		msg := resmsg.(*protocol.MsgError)
		return fmt.Errorf("Do login error: %s", msg.ErrTip.Value())
	}

	// 消息类型
	if resmsgty != protocol.MsgTypeLoginCheckLastestBill {
		// 不支持的消息类型
		wsconn.Close()
		return fmt.Errorf("Unsupported message reply type = %d", resmsgty)
	}

	// 登录成功
	billmsg := resmsg.(*protocol.MsgLoginCheckLastestBill)

	// 检查协议
	if uint16(billmsg.ProtocolVersion) > protocol.LatestProtocolVersion {
		wsconn.Close()
		return fmt.Errorf("You need to upgrade your client software, version %d => %d", protocol.LatestProtocolVersion, billmsg.ProtocolVersion)
	}
	if uint16(billmsg.ProtocolVersion) < protocol.LatestProtocolVersion {
		wsconn.Close()
		return fmt.Errorf("Your servicer does not support the protocol version %d of your client software", protocol.LatestProtocolVersion)
	}

	// 创建 upstreamChannelSide
	csobj := chanpay.NewChannelSideByConn(wsconn)
	csobj.ChannelId = c.selfAddr.ChannelId
	csobj.ChannelInfo = c.chanInfo
	csobj.OurAddress = c.selfAddr.Address
	ourIsLeft := c.selfAddr.Address.Equal(c.chanInfo.LeftAddress)
	if ourIsLeft {
		csobj.RemoteAddress = c.chanInfo.RightAddress
	} else {
		csobj.RemoteAddress = c.chanInfo.LeftAddress
	}
	if billmsg.BillIsExistent.Check() {
		csobj.LatestReconciliationBalanceBill = billmsg.LastBill
	}

	// 通道端
	c.servicerStreamSide = chanpay.NewRelayPayNodeConnect(c.selfAddr.ServicerName.Value(), csobj.ChannelId, ourIsLeft, csobj)

	// 心跳保活
	go c.StartHeartbeat()

	// 成功
	return nil
}
