package payroutes

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/core/fields"
	"io/ioutil"
	"path"
	"strings"
)

const (
	NodeRoutesDataFileNameOfState = "state.dat"
	NodeRoutesDataFileNameOfNodes = "nodes.dat"
	NodeRoutesDataFileNameOfGraph = "graph.dat"
)

/**
 * 从磁盘读取节点及关系表
 */
func (p *RoutingManager) LoadAllNodesAndRelationshipFormDisk(datadir string, datanodes, datagraph *[]byte) error {
	// locking
	p.nodeUpdateLock.Lock()
	defer p.nodeUpdateLock.Unlock()

	var e error

	// Read status
	statname := path.Join(datadir, NodeRoutesDataFileNameOfState)
	numbt, e := ioutil.ReadFile(statname)
	if e == nil && len(numbt) == 4 {
		// Resolution status
		e = p.RebuildNodesAndRelationshipUnsafe(NodeRoutesDataFileNameOfState, numbt)
		if e != nil {
			return e
		}
	}

	// Read nodes and relationships
	if p.nodeUpdateLastestPageNum > 0 {

		// Read all nodes
		nodesfn := path.Join(datadir, NodeRoutesDataFileNameOfNodes)
		nodesbts, e := ioutil.ReadFile(nodesfn)
		if e == nil && nodesbts != nil {
			// Resolve all nodes
			e = p.RebuildNodesAndRelationshipUnsafe(NodeRoutesDataFileNameOfNodes, nodesbts)
			if e != nil {
				return e
			}
			*datanodes = nodesbts // Copy out for use
		}

		// Read all relationships
		graphfn := path.Join(datadir, NodeRoutesDataFileNameOfGraph)
		graphbts, e := ioutil.ReadFile(graphfn)
		if e == nil && graphbts != nil {
			// Resolve all relationships
			e = p.RebuildNodesAndRelationshipUnsafe(NodeRoutesDataFileNameOfGraph, graphbts)
			if e != nil {
				return e
			}
			*datagraph = graphbts // Copy out for use
		}

	}

	return nil
}

/**
 * 重建节点与关系数据
 */
func (p *RoutingManager) RebuildNodesAndRelationshipUnsafe(fnamety string, conbts []byte) error {
	var e error
	if fnamety == NodeRoutesDataFileNameOfState {

		// state
		var curpagenum uint32 = 0
		curpagenum = binary.BigEndian.Uint32(conbts)
		p.nodeUpdateLastestPageNum = curpagenum

	} else if fnamety == NodeRoutesDataFileNameOfNodes {

		// node
		// Resolve all nodes
		var fseek uint32 = 0
		var nodeById = make(map[uint32]*PayRelayNode)   // node
		var nodeByName = make(map[string]*PayRelayNode) // node
		for {
			var node = &PayRelayNode{}
			fseek, e = node.Parse(conbts, fseek)
			if e != nil {
				break // All parsing completed
			}
			nodeById[uint32(node.ID)] = node
			nodeByName[strings.ToLower(node.IdentificationName.Value())] = node // ignore case
		}
		p.nodeById = nodeById
		p.nodeByName = nodeByName // Read complete

	} else if fnamety == NodeRoutesDataFileNameOfGraph {

		// relationship
		// Resolve all relationships
		var graphDatas = make([]*ChannelRelationship, 0, len(conbts)/8) // Relation table
		for i := 0; i+7 < len(conbts); i += 8 {
			n1 := binary.BigEndian.Uint32(conbts[i : i+4])
			n2 := binary.BigEndian.Uint32(conbts[i+4 : i+8])
			graphDatas = append(graphDatas, &ChannelRelationship{
				fields.VarUint4(n1), fields.VarUint4(n2),
			})
		}
		p.graphDatas = graphDatas // Read complete

	}

	// All complete
	return nil
}

/**
 * 将所有节点及关系表刷到磁盘永久保存
 */
func (p *RoutingManager) FlushAllNodesAndRelationshipToDiskUnsafe(datadir string, datanodes, datagraph *[]byte) error {

	// Judge whether there is data
	if p.nodeUpdateLastestPageNum == 0 {
		return nil
	}

	// Save node table
	var nodesbuf = bytes.NewBuffer(nil)
	for _, v := range p.nodeById {
		nbts, _ := v.Serialize()
		nodesbuf.Write(nbts)
	}
	*datanodes = nodesbuf.Bytes()

	// Save relationship table
	var graphbuf = bytes.NewBuffer(nil)
	for _, v := range p.graphDatas {
		nbts, _ := v.Serialize()
		graphbuf.Write(nbts)
	}
	*datagraph = graphbuf.Bytes()

	// Save status
	statbts := make([]byte, 4)
	binary.BigEndian.PutUint32(statbts, p.nodeUpdateLastestPageNum)

	// Write to disk
	return p.ForceWriteAllNodesAndRelationshipToDiskUnsafe(datadir, statbts, nodesbuf.Bytes(), graphbuf.Bytes())
}

/**
 * 保存到磁盘永久保存
 */
func (p *RoutingManager) ForceWriteAllNodesAndRelationshipToDiskUnsafe(datadir string, statbts, datanodes, datagraph []byte) error {

	var e error

	// Judge whether there is data
	if p.nodeUpdateLastestPageNum == 0 {
		return nil
	}

	// Save node table
	nodesfn := path.Join(datadir, NodeRoutesDataFileNameOfNodes)
	e = ioutil.WriteFile(nodesfn, datanodes, 0777)
	if e != nil {
		return e
	}

	// Save relationship table
	graphfn := path.Join(datadir, NodeRoutesDataFileNameOfGraph)
	e = ioutil.WriteFile(graphfn, datagraph, 0777)
	if e != nil {
		return e
	}

	// Save status
	statname := path.Join(datadir, NodeRoutesDataFileNameOfState)
	e = ioutil.WriteFile(statname, statbts, 0777)
	if e != nil {
		return e
	}

	// All saved
	return nil
}
