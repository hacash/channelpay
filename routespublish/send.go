package routespublish

import (
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
)

// Send all nodes
func (p *PayRoutesPublish) SendAllNodesData(ws *websocket.Conn) {
	msgobj := &protocol.MsgPayRouteResponseServiceNodes{
		LastestUpdatePageNum: fields.VarUint4(p.routingManager.GetUpdateLastestPageNum()),
		AllNodesBytes:        fields.CreateStringMax16777215(string(p.dataAllNodes)),
	}
	// send out
	protocol.SendMsg(ws, msgobj)
}

// Send all service node information
func (p *PayRoutesPublish) SendAllGraphData(ws *websocket.Conn) {
	msgobj := &protocol.MsgPayRouteResponseNodeRelationship{
		AllRelationships: fields.CreateStringMax16777215(string(p.dataAllGraph)),
	}
	// send out
	protocol.SendMsg(ws, msgobj)
}

// Send updates
func (p *PayRoutesPublish) SendTargetUpdateData(ws *websocket.Conn, queryPageNum uint32) {

	msgobj := &protocol.MsgPayRouteResponseUpdates{
		DataStatus:            2, // 超出了
		AllUpdatesOfJsonBytes: fields.CreateStringMax16777215(""),
	}

	// read file
	data, e := p.ReadUpdateLogFile(queryPageNum)
	if e != nil {
		protocol.SendMsg(ws, msgobj) // Exceeded
		return
	}

	// send files
	msgobj.DataStatus = 1 // Normal reading
	msgobj.AllUpdatesOfJsonBytes = fields.CreateStringMax16777215(string(data))
	protocol.SendMsg(ws, msgobj)
}
