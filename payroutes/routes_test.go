package payroutes

import (
	"fmt"
	"github.com/hacash/core/fields"
	"strings"
	"testing"
	"time"
)

func Test_t1(t *testing.T) {

	// 测试路由查询
	mng := RoutingManager{
		nodeById:   make(map[uint32]*PayRelayNode, 0),
		nodeByName: make(map[string]*PayRelayNode, 0),
		graphDatas: make([]*ChannelRelationship, 0),
	}

	// 添加十个节点
	overdueTime := time.Now().Unix() + 99999
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("node%d", i)
		node := &PayRelayNode{
			ID:                 fields.VarUint4(i),
			IdentificationName: fields.CreateStringMax255(name),
			OverdueTime:        fields.BlockTxTimestamp(overdueTime),
		}
		mng.nodeById[uint32(i)] = node
		mng.nodeByName[name] = node
	}

	// 添加关系表
	createRelationship := func(ships [][]uint32) []*ChannelRelationship {
		res := make([]*ChannelRelationship, len(ships))
		for i, v := range ships {
			res[i] = &ChannelRelationship{
				LeftNodeID:  fields.VarUint4(v[0]),
				RightNodeID: fields.VarUint4(v[1]),
			}
		}
		return res
	}
	mng.graphDatas = createRelationship([][]uint32{
		{1, 2},
		{1, 3},
		{1, 4},
		{4, 5},
		{5, 6},
		{3, 10},
		{6, 10},
		{4, 10},
	})

	// 测试路由查找
	pathnodes, e := mng.SearchNodePath("node1", "node5")
	if e != nil {
		fmt.Println(e.Error())
	} else {
		fmt.Printf("%d pathnodes:\n", len(pathnodes))
		for i, one := range pathnodes {
			nodens := make([]string, len(one))
			for a, v := range one {
				nodens[a] = v.IdentificationName.Value()
			}
			fmt.Printf("%d: %s\n", i+1, strings.Join(nodens, " -> "))
		}
	}

}
