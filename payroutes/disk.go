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
	// 锁定
	p.nodeUpdateLock.Lock()
	defer p.nodeUpdateLock.Unlock()

	var e error

	// 读取状态
	statname := path.Join(datadir, NodeRoutesDataFileNameOfState)
	numbt, e := ioutil.ReadFile(statname)
	if e == nil && len(numbt) == 4 {
		// 解析状态
		e = p.RebuildNodesAndRelationshipUnsafe(NodeRoutesDataFileNameOfState, numbt)
		if e != nil {
			return e
		}
	}

	// 读取节点及关系
	if p.nodeUpdateLastestPageNum > 0 {

		// 读取全部节点
		nodesfn := path.Join(datadir, NodeRoutesDataFileNameOfNodes)
		nodesbts, e := ioutil.ReadFile(nodesfn)
		if e == nil && nodesbts != nil {
			// 解析全部节点
			e = p.RebuildNodesAndRelationshipUnsafe(NodeRoutesDataFileNameOfNodes, nodesbts)
			if e != nil {
				return e
			}
			*datanodes = nodesbts // 拷贝出去使用
		}

		// 读取全部关系
		graphfn := path.Join(datadir, NodeRoutesDataFileNameOfGraph)
		graphbts, e := ioutil.ReadFile(graphfn)
		if e == nil && graphbts != nil {
			// 解析全部关系
			e = p.RebuildNodesAndRelationshipUnsafe(NodeRoutesDataFileNameOfGraph, graphbts)
			if e != nil {
				return e
			}
			*datagraph = graphbts // 拷贝出去使用
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

		// 状态
		var curpagenum uint32 = 0
		curpagenum = binary.BigEndian.Uint32(conbts)
		p.nodeUpdateLastestPageNum = curpagenum

	} else if fnamety == NodeRoutesDataFileNameOfNodes {

		// 节点
		// 解析全部节点
		var fseek uint32 = 0
		var nodeById = make(map[uint32]*PayRelayNode)   // 节点
		var nodeByName = make(map[string]*PayRelayNode) // 节点
		for {
			var node = &PayRelayNode{}
			fseek, e = node.Parse(conbts, fseek)
			if e != nil {
				break // 全部解析完毕
			}
			nodeById[uint32(node.ID)] = node
			nodeByName[strings.ToLower(node.IdentificationName.Value())] = node // 忽略大小写
		}
		p.nodeById = nodeById
		p.nodeByName = nodeByName // 读取完毕

	} else if fnamety == NodeRoutesDataFileNameOfGraph {

		// 关系
		// 解析全部关系
		var graphDatas = make([]*ChannelRelationship, 0, len(conbts)/8) // 关系表
		for i := 0; i+7 < len(conbts); i += 8 {
			n1 := binary.BigEndian.Uint32(conbts[i : i+4])
			n2 := binary.BigEndian.Uint32(conbts[i+4 : i+8])
			graphDatas = append(graphDatas, &ChannelRelationship{
				fields.VarUint4(n1), fields.VarUint4(n2),
			})
		}
		p.graphDatas = graphDatas // 读取完毕

	}

	// 全部完成
	return nil
}

/**
 * 将所有节点及关系表刷到磁盘永久保存
 */
func (p *RoutingManager) FlushAllNodesAndRelationshipToDiskUnsafe(datadir string, datanodes, datagraph *[]byte) error {

	// 判断有无数据
	if p.nodeUpdateLastestPageNum == 0 {
		return nil
	}

	// 保存节点表
	var nodesbuf = bytes.NewBuffer(nil)
	for _, v := range p.nodeById {
		nbts, _ := v.Serialize()
		nodesbuf.Write(nbts)
	}
	*datanodes = nodesbuf.Bytes()

	// 保存关系表
	var graphbuf = bytes.NewBuffer(nil)
	for _, v := range p.graphDatas {
		nbts, _ := v.Serialize()
		graphbuf.Write(nbts)
	}
	*datagraph = graphbuf.Bytes()

	// 保存状态
	statbts := make([]byte, 4)
	binary.BigEndian.PutUint32(statbts, p.nodeUpdateLastestPageNum)

	// 写入磁盘
	return p.ForceWriteAllNodesAndRelationshipToDiskUnsafe(datadir, statbts, nodesbuf.Bytes(), graphbuf.Bytes())
}

/**
 * 保存到磁盘永久保存
 */
func (p *RoutingManager) ForceWriteAllNodesAndRelationshipToDiskUnsafe(datadir string, statbts, datanodes, datagraph []byte) error {

	var e error

	// 判断有无数据
	if p.nodeUpdateLastestPageNum == 0 {
		return nil
	}

	// 保存节点表
	nodesfn := path.Join(datadir, NodeRoutesDataFileNameOfNodes)
	e = ioutil.WriteFile(nodesfn, datanodes, 0777)
	if e != nil {
		return e
	}

	// 保存关系表
	graphfn := path.Join(datadir, NodeRoutesDataFileNameOfGraph)
	e = ioutil.WriteFile(graphfn, datagraph, 0777)
	if e != nil {
		return e
	}

	// 保存状态
	statname := path.Join(datadir, NodeRoutesDataFileNameOfState)
	e = ioutil.WriteFile(statname, statbts, 0777)
	if e != nil {
		return e
	}

	// 全部保存完毕
	return nil
}
