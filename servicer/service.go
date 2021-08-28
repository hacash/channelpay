package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/payroutes"
	"github.com/hacash/channelpay/servicer/datasources"
	"sync"
)

type Servicer struct {
	config *ServicerConfig

	// 数据接口
	billstore datasources.DataSourceOfBalanceBill
	chanset   datasources.DataSourceOfServicerPayChannelSetup

	// 客户连接池
	customerChgLock sync.Mutex
	customers       map[string]*Customer

	// 路由管理器
	payRouteMng *payroutes.RoutingManager
}

func NewServicer(cnf *ServicerConfig) *Servicer {

	ser := &Servicer{
		config:      cnf,
		customers:   make(map[string]*Customer, 0),
		payRouteMng: payroutes.NewRoutingManager(),
	}

	return ser
}

// 启动
func (s *Servicer) Start() {

	// 从本地磁盘读取路由
	var d1 []byte
	var d2 []byte
	e := s.payRouteMng.LoadAllNodesAndRelationshipFormDisk(s.config.RoutesSourceDataDir, &d1, &d2)
	if e != nil {
		fmt.Println(e)
	}

	// 从远程加载路由
	s.checkInitLoadRoutes()

	go s.loop()

	go s.startListen()

}

// 设置数据来源接口
func (s *Servicer) SetDataSource(
	billstore datasources.DataSourceOfBalanceBill,
	chanset datasources.DataSourceOfServicerPayChannelSetup,
) {
	s.billstore = billstore
	s.chanset = chanset
}
