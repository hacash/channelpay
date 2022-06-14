package payroutes

import (
	"strings"
	"sync"
)

/**
 * 支付路由
 */

type RoutingManager struct {

	// data
	nodeUpdateLock           sync.Mutex               // Data locking
	nodeUpdateLastestPageNum uint32                   // Latest data update flip
	nodeById                 map[uint32]*PayRelayNode // node
	nodeByName               map[string]*PayRelayNode // Node name, case insensitive
	graphDatas               []*ChannelRelationship   // Relation table

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
