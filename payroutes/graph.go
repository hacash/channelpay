package payroutes

import (
	"encoding/binary"
	"github.com/hacash/core/fields"
)

/**
 * 通道连通关系
 */
type ChannelRelationship struct {
	LeftNodeID  fields.VarUint4 // Node 1
	RightNodeID fields.VarUint4 // Node 2

}

func (c ChannelRelationship) Serialize() ([]byte, error) {
	var bt = make([]byte, 8)
	binary.BigEndian.PutUint32(bt[0:4], uint32(c.LeftNodeID))
	binary.BigEndian.PutUint32(bt[4:8], uint32(c.RightNodeID))
	return bt, nil
}
