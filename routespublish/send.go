package routespublish

import (
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/core/fields"
	"github.com/hacash/node/websocket"
)

// 发送全部节点
func (p *PayRoutesPublish) SendAllNodesData(ws *websocket.Conn) {
	msgobj := &protocol.MsgPayRouteResponseServiceNodes{
		LastestUpdatePageNum: fields.VarUint4(p.routingManager.GetUpdateLastestPageNum()),
		AllNodesBytes:        fields.CreateStringMax16777215(string(p.dataAllNodes)),
	}
	// 发送
	protocol.SendMsg(ws, msgobj)
}

// 发送全部服务节点信息
func (p *PayRoutesPublish) SendAllGraphData(ws *websocket.Conn) {
	msgobj := &protocol.MsgPayRouteResponseNodeRelationship{
		AllRelationships: fields.CreateStringMax16777215(string(p.dataAllGraph)),
	}
	// 发送
	protocol.SendMsg(ws, msgobj)
}

// 发送更新
func (p *PayRoutesPublish) SendTargetUpdateData(ws *websocket.Conn, queryPageNum uint32) {

	msgobj := &protocol.MsgPayRouteResponseUpdates{
		DataStatus:            2, // 超出了
		AllUpdatesOfJsonBytes: fields.CreateStringMax16777215(""),
	}

	// 读取文件
	data, e := p.ReadUpdateLogFile(queryPageNum)
	if e != nil {
		protocol.SendMsg(ws, msgobj) // 超出了
		return
	}

	// 发送文件
	msgobj.DataStatus = 1 // 正常读取
	msgobj.AllUpdatesOfJsonBytes = fields.CreateStringMax16777215(string(data))
	protocol.SendMsg(ws, msgobj)
}
