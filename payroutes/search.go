package payroutes

import (
	"fmt"
	"sort"
	"time"
)

type nodePathIdList [][]uint32

func (n nodePathIdList) Len() int {
	return len(n)
}

func (n nodePathIdList) Less(i, j int) bool {
	//fmt.Println(i, j, n[i] < n[j], n)
	return len(n[i]) < len(n[j])
}

func (n nodePathIdList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

/**
 * 搜索路径
 * 返回路径的节点路径表
 * 只通过关系查找，不管容量
 */
func (r *RoutingManager) SearchNodePath(startName, targetName string) ([][]*PayRelayNode, error) {
	if startName == targetName {
		return nil, fmt.Errorf("startName and targetName is cannot same")
	}
	var respath = make([][]uint32, 0)
	var respathnodes = make([][]*PayRelayNode, 0)
	// 解析id
	startNode := r.nodeByName[startName]
	targetNode := r.nodeByName[targetName]
	if startNode == nil {
		return nil, fmt.Errorf("Start node name <%s> not find", startName)
	}
	if targetNode == nil {
		return nil, fmt.Errorf("Target node name <%s> not find", targetName)
	}
	// 开始搜索
	r.doSearchPath(0, uint32(startNode.ID), uint32(targetNode.ID), []uint32{}, &respath)

	// 按路径长短排序
	sort.Sort(nodePathIdList(respath))

	// 替换为节点
	timeNow := time.Now().Unix()
	for _, idlist := range respath {
		npath := make([]*PayRelayNode, len(idlist))
		for i, id := range idlist {
			p := r.nodeById[id]
			if p == nil || int64(p.OverdueTime) < timeNow {
				npath = nil // 剔除不存在或服务过期的
				break
			} else {
				npath[i] = p
			}
		}
		if npath == nil {
			continue
		}
		respathnodes = append(respathnodes, npath)
	}

	// 返回
	return respathnodes, nil
}

// 查询，且递归查询
func (r *RoutingManager) doSearchPath(recursion, start, target uint32, prefixPath []uint32, respath *[][]uint32) {
	recursion += 1
	if recursion >= 8 {
		return // 递归最多 8 层
	}
	// 左侧，第一步查找
	nexts := r.findOutRelationship(start, prefixPath)
	if isContainUint32(nexts, target) {
		// 找到了，评出完整路径
		onepath := make([]uint32, 0)
		for _, v := range prefixPath {
			onepath = append(onepath, v)
		}
		onepath = append(onepath, start, target)
		*respath = append(*respath, onepath)
	}
	//// 右侧，第二步查找
	//prevs := r.findOutRelationship(target, subfixPath)
	// 递归查找
	for _, v := range nexts {
		if v != target && v != start && !isContainUint32(prefixPath, v) {
			newPrefixPath := make([]uint32, len(prefixPath))
			copy(newPrefixPath, prefixPath)
			newPrefixPath = append(newPrefixPath, start)
			// 递归查找
			r.doSearchPath(recursion, v, target, newPrefixPath, respath)
		}
	}

	// 完成查询
}

func (r *RoutingManager) findOutRelationship(myid uint32, avoidloopback []uint32) []uint32 {

	reslist := make([]uint32, 0)
	for _, v := range r.graphDatas {
		id1 := uint32(v.LeftNodeID)
		id2 := uint32(v.RightNodeID)
		if id1 == myid && id2 != myid {
			if !isContainUint32(avoidloopback, id2) {
				reslist = append(reslist, id2)
			}
		} else if id2 == myid && id1 != myid {
			if !isContainUint32(avoidloopback, id1) {
				reslist = append(reslist, id1)
			}
		}
	}
	return reslist
}

func isContainUint32(items []uint32, item uint32) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}
