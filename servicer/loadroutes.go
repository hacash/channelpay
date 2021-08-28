package servicer

import (
	"encoding/binary"
	"fmt"
	"github.com/hacash/channelpay/payroutes"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
)

/**
 * 加载最新路由数据
 */

// 启动事件循环
func (s *Servicer) dialToPublisher() (*websocket.Conn, error) {
	fmt.Println("dial to routes publisher:", s.config.LoadRoutesUrl)
	return websocket.Dial(s.config.LoadRoutesUrl, "ws", "http://127.0.0.1/")
}

// 加载最新更新
func (s *Servicer) LoadRoutesUpdate() {

	// 首次下载全部数据
	wsConn, e := s.dialToPublisher()
	if e != nil {
		fmt.Printf("LoadRoutesUpdate Dial to %s Error: %s \n", s.config.LoadRoutesUrl, e.Error())
		return
	}
	// 超时关闭
	// 锁定
	s.payRouteMng.UpdateLock()
	defer s.payRouteMng.UpdateUnlock()

	// 循环请求
	queryNum := s.payRouteMng.GetUpdateLastestPageNum()
	var canFlush = false // 是否刷新磁盘

	fmt.Printf("loadRoutesUpdate target page %d.\n", queryNum+1)

	for {
		var e error
		queryNum++
		// 请求更新
		msg1 := protocol.MsgPayRouteRequestUpdates{
			QueryPageNum: fields.VarUint4(queryNum),
		}
		e = protocol.SendMsg(wsConn, &msg1)
		if e != nil {
			fmt.Printf("checkInitLoadRoutes MsgPayRouteRequestServiceNodes Error: %s \n", e.Error())
			return
		}
		// 应答
		msgres1, _, e := protocol.ReceiveMsg(wsConn)
		obj1 := msgres1.(*protocol.MsgPayRouteResponseUpdates)
		if obj1 == nil {
			return // 错误消息
		}
		if obj1.DataStatus != 1 {
			// 超过了
			break
		}
		fmt.Printf("got routes data page: %d.\n", queryNum)
		// 更新
		s.payRouteMng.ForceUpdataNodesAndRelationshipByJsonBytesUnsafe([]byte(obj1.AllUpdatesOfJsonBytes.Value()), queryNum)
		canFlush = true
	}
	// 保存
	if canFlush {

		fmt.Printf("flush update routes data to disk.\n")
		var d1 []byte
		var d2 []byte
		s.payRouteMng.FlushAllNodesAndRelationshipToDiskUnsafe(s.config.RoutesSourceDataDir, &d1, &d2)
	} else {
		fmt.Printf("routes data is fresh.\n")
	}

}

// 检查或初始化路由加载
func (s *Servicer) checkInitLoadRoutes() {

	curnum := s.payRouteMng.GetUpdateLastestPageNum()
	if curnum > 0 {
		// 加载更新
		s.LoadRoutesUpdate()
		return
	}

	// 首次下载全部数据
	wsConn, e := s.dialToPublisher()
	if e != nil {
		fmt.Printf("checkInitLoadRoutes Dial to %s Error: %s \n", s.config.LoadRoutesUrl, e.Error())
		return
	}
	// 超时关闭
	// 请求全部节点数据
	msg1 := protocol.MsgPayRouteRequestServiceNodes{}
	e = protocol.SendMsg(wsConn, &msg1)
	if e != nil {
		fmt.Printf("checkInitLoadRoutes MsgPayRouteRequestServiceNodes Error: %s \n", e.Error())
		return
	}
	// 应答
	msgres1, _, e := protocol.ReceiveMsg(wsConn)
	// 请求关系表
	msg2 := protocol.MsgPayRouteRequestNodeRelationship{}
	e = protocol.SendMsg(wsConn, &msg2)
	if e != nil {
		fmt.Printf("checkInitLoadRoutes MsgPayRouteRequestNodeRelationship Error: %s \n", e.Error())
		return
	}
	// 应答
	msgres2, _, e := protocol.ReceiveMsg(wsConn)

	// 全部请求完毕，关闭
	protocol.SendMsg(wsConn, &protocol.MsgPayRouteEndClose{})
	wsConn.Close()

	// 保存
	// 锁定
	s.payRouteMng.UpdateLock()
	defer s.payRouteMng.UpdateUnlock()

	// 建立
	var d0 []byte
	var d1 []byte
	var d2 []byte
	obj1 := msgres1.(*protocol.MsgPayRouteResponseServiceNodes)
	if obj1 != nil {
		// 最新
		numbts := make([]byte, 4)
		binary.BigEndian.PutUint32(numbts, uint32(obj1.LastestUpdatePageNum))
		s.payRouteMng.RebuildNodesAndRelationshipUnsafe(payroutes.NodeRoutesDataFileNameOfState, numbts)
		d0 = numbts
		// 所有节点
		d1 = []byte(obj1.AllNodesBytes.Value())
		s.payRouteMng.RebuildNodesAndRelationshipUnsafe(payroutes.NodeRoutesDataFileNameOfNodes, d1)
		fmt.Printf("Got Routes Nodes len %d lastest page %d.\n",
			obj1.AllNodesBytes.Len, obj1.LastestUpdatePageNum)
	}
	obj2 := msgres2.(*protocol.MsgPayRouteResponseNodeRelationship)
	if obj2 != nil {
		// 全部关系
		d2 = []byte(obj2.AllRelationships.Value())
		s.payRouteMng.RebuildNodesAndRelationshipUnsafe(payroutes.NodeRoutesDataFileNameOfGraph, d2)
		fmt.Printf("Got Routes Relationship len %d.\n", obj2.AllRelationships.Len)
	}

	// 全部建立完成，保存到磁盘
	s.payRouteMng.ForceWriteAllNodesAndRelationshipToDiskUnsafe(s.config.RoutesSourceDataDir, d0, d1, d2)
	// all ok
}
