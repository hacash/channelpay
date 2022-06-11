package routespublish

import (
	"fmt"
	"github.com/hacash/channelpay/payroutes"
)

/**
 * 路由数据下发
 */

type PayRoutesPublish struct {
	config *PayRoutesPublishConfig

	// Data cache
	dataAllNodes []byte
	dataAllGraph []byte

	// Administration
	routingManager *payroutes.RoutingManager
}

//
func NewPayRoutesPublish(config *PayRoutesPublishConfig) *PayRoutesPublish {

	rtmng := payroutes.NewRoutingManager()

	return &PayRoutesPublish{
		config:         config,
		dataAllNodes:   nil,
		dataAllGraph:   nil,
		routingManager: rtmng,
	}
}

// start
func (p *PayRoutesPublish) Start() {

	// Read configuration from directory
	e := p.routingManager.LoadAllNodesAndRelationshipFormDisk(
		p.config.DataSourceDir, &p.dataAllNodes, &p.dataAllGraph,
	)
	if e != nil {
		fmt.Println("PayRoutesPublish.Start() Error:", e.Error())
	} else {
		fmt.Printf("current update log file num is %d.\n", p.routingManager.GetUpdateLastestPageNum())
	}

	// Read update log from disk
	p.DoUpdateByReadLogFile()

	// event processing 
	go p.loop()

	// monitor
	go p.listen(p.config.WssListenPort)

}
