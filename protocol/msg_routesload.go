package protocol

import (
	"bytes"
	"github.com/hacash/core/fields"
)

/**
 * 支付节点路由下载同步
 */

// Complete and close
type MsgPayRouteEndClose struct {
}

func (m MsgPayRouteEndClose) Type() uint8 {
	return MsgTypePayRouteEndClose
}

func (m MsgPayRouteEndClose) Size() uint32 {
	return 0
}

func (m *MsgPayRouteEndClose) Parse(buf []byte, seek uint32) (uint32, error) {
	return seek, nil
}

func (m MsgPayRouteEndClose) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	// ok
	return buf.Bytes(), nil
}

func (m MsgPayRouteEndClose) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

// Request list of all nodes
type MsgPayRouteRequestServiceNodes struct {
}

func (m MsgPayRouteRequestServiceNodes) Type() uint8 {
	return MsgTypePayRouteRequestServiceNodes
}

func (m MsgPayRouteRequestServiceNodes) Size() uint32 {
	return 0
}

func (m *MsgPayRouteRequestServiceNodes) Parse(buf []byte, seek uint32) (uint32, error) {
	return seek, nil
}

func (m MsgPayRouteRequestServiceNodes) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	// ok
	return buf.Bytes(), nil
}

func (m MsgPayRouteRequestServiceNodes) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

// Respond to all nodes
type MsgPayRouteResponseServiceNodes struct {
	LastestUpdatePageNum fields.VarUint4
	AllNodesBytes        fields.StringMax16777215
}

func (m MsgPayRouteResponseServiceNodes) Type() uint8 {
	return MsgTypePayRouteResponseServiceNodes
}

func (m MsgPayRouteResponseServiceNodes) Size() uint32 {
	return m.LastestUpdatePageNum.Size() + m.AllNodesBytes.Size()
}

func (m *MsgPayRouteResponseServiceNodes) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.LastestUpdatePageNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.AllNodesBytes.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgPayRouteResponseServiceNodes) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b1, e := m.LastestUpdatePageNum.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	b2, e := m.AllNodesBytes.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b2)
	// ok
	return buf.Bytes(), nil
}

func (m MsgPayRouteResponseServiceNodes) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

// Request all relationships
type MsgPayRouteRequestNodeRelationship struct {
}

func (m MsgPayRouteRequestNodeRelationship) Type() uint8 {
	return MsgTypePayRouteRequestNodeRelationship
}

func (m MsgPayRouteRequestNodeRelationship) Size() uint32 {
	return 0
}

func (m *MsgPayRouteRequestNodeRelationship) Parse(buf []byte, seek uint32) (uint32, error) {
	return seek, nil
}

func (m MsgPayRouteRequestNodeRelationship) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	// ok
	return buf.Bytes(), nil
}

func (m MsgPayRouteRequestNodeRelationship) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

// Respond to all relationships
type MsgPayRouteResponseNodeRelationship struct {
	AllRelationships fields.StringMax16777215
}

func (m MsgPayRouteResponseNodeRelationship) Type() uint8 {
	return MsgTypePayRouteResponseNodeRelationship
}

func (m MsgPayRouteResponseNodeRelationship) Size() uint32 {
	return m.AllRelationships.Size()
}

func (m *MsgPayRouteResponseNodeRelationship) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.AllRelationships.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgPayRouteResponseNodeRelationship) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b1, e := m.AllRelationships.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

func (m MsgPayRouteResponseNodeRelationship) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

// Request route modification
type MsgPayRouteRequestUpdates struct {
	QueryPageNum fields.VarUint4 // Requested pages
}

func (m MsgPayRouteRequestUpdates) Type() uint8 {
	return MsgTypePayRouteRequestUpdates
}

func (m MsgPayRouteRequestUpdates) Size() uint32 {
	return m.QueryPageNum.Size()
}

func (m *MsgPayRouteRequestUpdates) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.QueryPageNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgPayRouteRequestUpdates) Serialize() ([]byte, error) {
	var e error
	var bt []byte
	buf := bytes.NewBuffer(nil)
	bt, e = m.QueryPageNum.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	// ok
	return buf.Bytes(), nil
}

func (m MsgPayRouteRequestUpdates) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}

// Response route update
type MsgPayRouteResponseUpdates struct {
	DataStatus            fields.VarUint1 // 数据状态  0.错误 1.正常 2.已超出最新 4.没找到
	AllUpdatesOfJsonBytes fields.StringMax16777215
}

func (m MsgPayRouteResponseUpdates) Type() uint8 {
	return MsgTypePayRouteResponseUpdates
}

func (m MsgPayRouteResponseUpdates) Size() uint32 {
	return m.DataStatus.Size() + m.AllUpdatesOfJsonBytes.Size()
}

func (m *MsgPayRouteResponseUpdates) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error
	seek, e = m.DataStatus.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = m.AllUpdatesOfJsonBytes.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (m MsgPayRouteResponseUpdates) Serialize() ([]byte, error) {
	var e error
	var bt []byte
	buf := bytes.NewBuffer(nil)
	bt, e = m.DataStatus.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	bt, e = m.AllUpdatesOfJsonBytes.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(bt)
	// ok
	return buf.Bytes(), nil
}

func (m MsgPayRouteResponseUpdates) SerializeWithType() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{m.Type()})
	b1, e := m.Serialize()
	if e != nil {
		return nil, e
	}
	buf.Write(b1)
	// ok
	return buf.Bytes(), nil
}
