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

	// 数据缓存
	dataAllNodes []byte
	dataAllGraph []byte

	// 管理
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

// 开始
func (p *PayRoutesPublish) Start() {

	// 从目录读取配置
	e := p.routingManager.LoadAllNodesAndRelationshipFormDisk(
		p.config.DataSourceDir, &p.dataAllNodes, &p.dataAllGraph,
	)
	if e != nil {
		fmt.Println("PayRoutesPublish.Start() Error:", e.Error())
	} else {
		fmt.Printf("current update log file num is %d.\n", p.routingManager.GetUpdateLastestPageNum())
	}

	// 从磁盘读取更新日志
	p.DoUpdateByReadLogFile()

	// 事件处理
	go p.loop()

	// 监听
	go p.listen(p.config.WssListenPort)

}
