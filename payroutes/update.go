package payroutes

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hacash/core/fields"
	"strings"
)

/**
 * 通过 json 文件更新节点及关系
 */

/*

// Update file format

{
	nodes: {
		insert: [{
			id: ...
			...
		}],
		update: [{
			id: ...
			...
		}],
		delete: [1,...]
	},

	graph: {
		add: [
			[3, 4],
			...
		],
		del: [
			[1, 2],
			...
		]
	}

}

*/

/**
 * 从json 解析更新当前的节点和关系表
 */

func createNodeFromJsonVal(data []byte, old *PayRelayNode) *PayRelayNode {
	var newnode *PayRelayNode = nil
	if old != nil {
		newnode = old.Copy()
	} else {
		newnode = &PayRelayNode{}
	}
	// analysis
	// id
	id, _ := jsonparser.GetInt(data, "id")
	if id > 0 {
		newnode.ID = fields.VarUint4(id)
	}
	// country_code
	country_code, _ := jsonparser.GetString(data, "country_code")
	if len(country_code) > 0 {
		newnode.CountryCode = fields.Bytes2([]byte(country_code))
	}
	// identification_name
	identification_name, _ := jsonparser.GetString(data, "identification_name")
	if len(identification_name) > 0 {
		newnode.IdentificationName = fields.CreateStringMax255(identification_name)
	}
	// fee_min
	feemin, _ := jsonparser.GetString(data, "fee_min")
	if len(feemin) > 0 {
		if fee, e := fields.NewAmountFromFinString(feemin); e == nil {
			newnode.FeeMin = *fee
		}
	}
	// fee_ratio
	fee_ratio, _ := jsonparser.GetInt(data, "fee_ratio")
	if id > 0 {
		newnode.FeeRatio = fields.VarUint4(fee_ratio)
	}
	// fee_max
	feemax, _ := jsonparser.GetString(data, "fee_max")
	if len(feemax) > 0 {
		if fee, e := fields.NewAmountFromFinString(feemax); e == nil {
			newnode.FeeMax = *fee
		}
	}
	// gateway_1
	gateway_1, _ := jsonparser.GetString(data, "gateway_1")
	if len(gateway_1) > 0 {
		newnode.Gateway1 = fields.CreateStringMax255(gateway_1)
	}
	// gateway_2
	gateway_2, _ := jsonparser.GetString(data, "gateway_2")
	if len(gateway_2) > 0 {
		newnode.Gateway2 = fields.CreateStringMax255(gateway_2)
	}
	// overdue_time
	overdue_time, _ := jsonparser.GetInt(data, "overdue_time")
	if overdue_time > 0 {
		newnode.OverdueTime = fields.VarUint5(overdue_time)
	}
	// overdue_time
	register_time, _ := jsonparser.GetInt(data, "register_time")
	if register_time > 0 {
		newnode.RegisterTime = fields.VarUint5(register_time)
	}

	return newnode
}

// to update
func (p *RoutingManager) ForceUpdataNodesAndRelationshipByJsonBytesUnsafe(databytes []byte, filenum uint32) error {

	var createNodeIdFromJsonVal = func(data []byte) uint32 {
		id, _ := jsonparser.GetInt(data, "id")
		return uint32(id)
	}

	// node
	jnodeval, _, _, _ := jsonparser.Get(databytes, "nodes")
	if jnodeval != nil {

		// Node addition
		ists, iststy, _, _ := jsonparser.Get(jnodeval, "insert")
		if ists != nil && iststy == jsonparser.Array {
			jsonparser.ArrayEach(ists, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				//fmt.Println(jsonparser.Get(value, "url"))
				node := createNodeFromJsonVal(value, nil)
				if node != nil && node.ID > 0 && node.IdentificationName.Len > 0 {
					// Check repetition
					if _, has := p.nodeById[uint32(node.ID)]; has {
						fmt.Printf("ForceUpdataNodesAndRelationshipByJsonBytes Insert Error: node id <%d> already exists.\n", node.ID)
					} else {
						// Insert node
						p.nodeById[uint32(node.ID)] = node
						p.nodeByName[strings.ToLower(node.IdentificationName.Value())] = node
					}
				}
			})
		}

		// Modify node
		chgts, chgty, _, _ := jsonparser.Get(jnodeval, "update")
		if ists != nil && chgty == jsonparser.Array {
			jsonparser.ArrayEach(chgts, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				//fmt.Println(jsonparser.Get(value, "url"))
				nodeid := createNodeIdFromJsonVal(value)
				if nodeid > 0 {
					// Check repetition
					if oldnode, has := p.nodeById[uint32(nodeid)]; !has {
						fmt.Printf("ForceUpdataNodesAndRelationshipByJsonBytes Update Error: node id <%d> not find.\n", nodeid)
					} else {
						// Update node
						delete(p.nodeByName, strings.ToLower(oldnode.IdentificationName.Value()))
						newnode := createNodeFromJsonVal(value, oldnode) // 创建节点
						p.nodeById[uint32(newnode.ID)] = newnode
						p.nodeByName[strings.ToLower(newnode.IdentificationName.Value())] = newnode
					}
				}
			})
		}

		// Delete node
		delts, delty, _, _ := jsonparser.Get(jnodeval, "delete")
		if ists != nil && delty == jsonparser.Array {
			jsonparser.ArrayEach(delts, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				//fmt.Println(jsonparser.Get(value, "url"))
				nodeid := createNodeIdFromJsonVal(value)
				if nodeid > 0 {
					// Check repetition
					if oldnode, has := p.nodeById[uint32(nodeid)]; !has {
						fmt.Printf("ForceUpdataNodesAndRelationshipByJsonBytes Delete Error: node id <%d> not find.\n", nodeid)
					} else {
						// Delete node
						delete(p.nodeByName, strings.ToLower(oldnode.IdentificationName.Value()))
						delete(p.nodeById, uint32(oldnode.ID))
					}
				}
			})
		}

	}

	// relationship
	jgraphval, _, _, _ := jsonparser.Get(databytes, "graph")
	if jgraphval != nil {
		// New relation
		ists, iststy, _, _ := jsonparser.Get(jgraphval, "add")
		if ists != nil && iststy == jsonparser.Array {
			jsonparser.ArrayEach(ists, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				n1, _ := jsonparser.GetInt(value, "[0]")
				id1 := fields.VarUint4(n1)
				n2, _ := jsonparser.GetInt(value, "[1]")
				id2 := fields.VarUint4(n2)
				if id1 > 0 && id2 > 0 {
					// Add lookup
					hav := false
					for _, v := range p.graphDatas {
						if (v.LeftNodeID == id1 && v.RightNodeID == id2) ||
							(v.LeftNodeID == id2 && v.RightNodeID == id1) {
							hav = true // Already exists
							break
						}
					}
					if !hav {
						// increase
						p.graphDatas = append(p.graphDatas, &ChannelRelationship{
							LeftNodeID:  id1,
							RightNodeID: id2,
						})
					}
				}
			})
		}
		// Relationship deletion
		delts, delty, _, _ := jsonparser.Get(jgraphval, "del")
		if delts != nil && delty == jsonparser.Array {
			// del
			jsonparser.ArrayEach(ists, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				n1, _ := jsonparser.GetInt(value, "[0]")
				id1 := fields.VarUint4(n1)
				n2, _ := jsonparser.GetInt(value, "[1]")
				id2 := fields.VarUint4(n2)
				if id1 > 0 && id2 > 0 {
					// Del find
					var hav int = -1
					for i, v := range p.graphDatas {
						if (v.LeftNodeID == id1 && v.RightNodeID == id2) ||
							(v.LeftNodeID == id2 && v.RightNodeID == id1) {
							hav = i // Already exists
							break
						}
					}
					if hav > -1 {
						// Delete array elements
						p.graphDatas = append(p.graphDatas[:hav], p.graphDatas[hav+1:]...)
					}
				}
			})

		}
	}

	// Update latest page number
	p.nodeUpdateLastestPageNum = filenum
	return nil
}
