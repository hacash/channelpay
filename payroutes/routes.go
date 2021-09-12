package payroutes

import (
	"strings"
	"sync"
)

/**
 * 支付路由
 */

type RoutingManager struct {

	// 数据
	nodeUpdateLock           sync.Mutex               // 数据锁定
	nodeUpdateLastestPageNum uint32                   // 最新数据更新翻页
	nodeById                 map[uint32]*PayRelayNode // 节点
	nodeByName               map[string]*PayRelayNode // 节点名称，不区分大小写
	graphDatas               []*ChannelRelationship   // 关系表

}

func NewRoutingManager() *RoutingManager {
	return &RoutingManager{
		nodeUpdateLock:           sync.Mutex{},
		nodeUpdateLastestPageNum: 0,
		nodeById:                 make(map[uint32]*PayRelayNode),
		nodeByName:               make(map[string]*PayRelayNode),
		graphDatas:               make([]*ChannelRelationship, 0),
	}
}

func (r *RoutingManager) GetUpdateLastestPageNum() uint32 {
	return r.nodeUpdateLastestPageNum
}

//func (r *RoutingManager) SetUpdateLastestPageNum(n uint32) {
//	r.nodeUpdateLastestPageNum = n
//}

func (r *RoutingManager) UpdateLock() {
	r.nodeUpdateLock.Lock()
}

func (r *RoutingManager) UpdateUnlock() {
	r.nodeUpdateLock.Unlock()
}

// find node by name
func (r *RoutingManager) FindNodeByName(name string) *PayRelayNode {
	return r.nodeByName[strings.ToLower(name)]
}
func (r *RoutingManager) FindNodeById(id uint32) *PayRelayNode {
	return r.nodeById[id]
}
