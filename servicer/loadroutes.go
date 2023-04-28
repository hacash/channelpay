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

// Start event cycle
func (s *Servicer) dialToPublisher() (*websocket.Conn, error) {
	fmt.Println("dial to routes publisher:", s.config.LoadRoutesUrl)
	return websocket.Dial(s.config.LoadRoutesUrl, "ws", "http://127.0.0.1/")
}

// Load latest updates
func (s *Servicer) LoadRoutesUpdate() {

	// Download all data for the first time
	wsConn, e := s.dialToPublisher()
	if e != nil {
		fmt.Printf("LoadRoutesUpdate Dial to %s Error: %s \n", s.config.LoadRoutesUrl, e.Error())
		return
	}
	// Timeout shutdown
	// locking
	s.payRouteMng.UpdateLock()
	defer s.payRouteMng.UpdateUnlock()

	// Circular request
	queryNum := s.payRouteMng.GetUpdateLastestPageNum()
	var canFlush = false // Refresh disk

	fmt.Printf("loadRoutesUpdate target page %d.\n", queryNum+1)

	for {
		var e error
		queryNum++
		// Request update
		msg1 := protocol.MsgPayRouteRequestUpdates{
			QueryPageNum: fields.VarUint4(queryNum),
		}
		e = protocol.SendMsg(wsConn, &msg1)
		if e != nil {
			fmt.Printf("checkInitLoadRoutes MsgPayRouteRequestServiceNodes Error: %s \n", e.Error())
			return
		}
		// answer
		msgres1, _, e := protocol.ReceiveMsg(wsConn)
		obj1 := msgres1.(*protocol.MsgPayRouteResponseUpdates)
		if obj1 == nil {
			return // Error message
		}
		if obj1.DataStatus != 1 {
			// Exceeded
			break
		}
		fmt.Printf("got routes data page: %d.\n", queryNum)
		// to update
		s.payRouteMng.ForceUpdataNodesAndRelationshipByJsonBytesUnsafe([]byte(obj1.AllUpdatesOfJsonBytes.Value()), queryNum)
		canFlush = true
	}
	// preservation
	if canFlush {

		fmt.Printf("flush update routes data to disk.\n")
		var d1 []byte
		var d2 []byte
		s.payRouteMng.FlushAllNodesAndRelationshipToDiskUnsafe(s.config.RoutesSourceDataDir, &d1, &d2)
	} else {
		fmt.Printf("routes data is fresh.\n")
	}

}

// Check or initialize route loading
func (s *Servicer) checkInitLoadRoutes() {

	curnum := s.payRouteMng.GetUpdateLastestPageNum()
	if curnum > 0 {
		// Load updates
		s.LoadRoutesUpdate()
		return
	}

	// Download all data for the first time
	wsConn, e := s.dialToPublisher()
	if e != nil {
		fmt.Printf("checkInitLoadRoutes Dial to %s Error: %s \n", s.config.LoadRoutesUrl, e.Error())
		return
	}
	// Timeout shutdown
	// Request all node data
	msg1 := protocol.MsgPayRouteRequestServiceNodes{}
	e = protocol.SendMsg(wsConn, &msg1)
	if e != nil {
		fmt.Printf("checkInitLoadRoutes MsgPayRouteRequestServiceNodes Error: %s \n", e.Error())
		return
	}
	// answer
	msgres1, _, e := protocol.ReceiveMsg(wsConn)
	// Request relation table
	msg2 := protocol.MsgPayRouteRequestNodeRelationship{}
	e = protocol.SendMsg(wsConn, &msg2)
	if e != nil {
		fmt.Printf("checkInitLoadRoutes MsgPayRouteRequestNodeRelationship Error: %s \n", e.Error())
		return
	}
	// answer
	msgres2, _, e := protocol.ReceiveMsg(wsConn)

	// All requests completed, close
	protocol.SendMsg(wsConn, &protocol.MsgPayRouteEndClose{})
	wsConn.Close()

	// preservation
	// locking
	s.payRouteMng.UpdateLock()
	defer s.payRouteMng.UpdateUnlock()

	// establish
	var d0 []byte
	var d1 []byte
	var d2 []byte
	obj1 := msgres1.(*protocol.MsgPayRouteResponseServiceNodes)
	if obj1 != nil {
		// newest
		numbts := make([]byte, 4)
		binary.BigEndian.PutUint32(numbts, uint32(obj1.LastestUpdatePageNum))
		s.payRouteMng.RebuildNodesAndRelationshipUnsafe(payroutes.NodeRoutesDataFileNameOfState, numbts)
		d0 = numbts
		// All nodes
		d1 = []byte(obj1.AllNodesBytes.Value())
		s.payRouteMng.RebuildNodesAndRelationshipUnsafe(payroutes.NodeRoutesDataFileNameOfNodes, d1)
		fmt.Printf("Got Routes Nodes bytes(%d) lastest page %d.\n",
			obj1.AllNodesBytes.Len, obj1.LastestUpdatePageNum)
	}
	obj2 := msgres2.(*protocol.MsgPayRouteResponseNodeRelationship)
	if obj2 != nil {
		// All relationships
		d2 = []byte(obj2.AllRelationships.Value())
		s.payRouteMng.RebuildNodesAndRelationshipUnsafe(payroutes.NodeRoutesDataFileNameOfGraph, d2)
		fmt.Printf("Got Routes Relationship bytes(%d).\n", obj2.AllRelationships.Len)
	}

	// All are created and saved to disk
	s.payRouteMng.ForceWriteAllNodesAndRelationshipToDiskUnsafe(s.config.RoutesSourceDataDir, d0, d1, d2)
	// all ok
}
