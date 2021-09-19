package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/payroutes"
	"sync"
)

type Servicer struct {
	config *ServicerConfig

	// 数据接口
	billstore   chanpay.DataSourceOfBalanceBill             // 票据储存
	chanset     chanpay.DataSourceOfServicerPayChannelSetup // 通道配置
	signmachine chanpay.DataSourceOfSignatureMachine        // 签名机器

	// 客户连接池
	customerChgLock sync.RWMutex
	customers       map[uint64]*chanpay.Customer // 客户端

	// 结算通道连接池
	settlenoderChgLock sync.RWMutex
	settlenoder        map[string][]*chanpay.RelayPaySettleNoder // 结算通道

	// 路由管理器
	payRouteMng      *payroutes.RoutingManager
	localServiceNode *payroutes.PayRelayNode // 本地服务节点
}

func NewServicer(cnf *ServicerConfig) *Servicer {

	ser := &Servicer{
		config:      cnf,
		customers:   make(map[uint64]*chanpay.Customer, 0),
		settlenoder: make(map[string][]*chanpay.RelayPaySettleNoder, 0),
		payRouteMng: payroutes.NewRoutingManager(),
	}

	return ser
}

// 启动
func (s *Servicer) Start() {

	var e error

	// 初始化结算通道
	s.setupRelaySettlementChannelDataSettings()

	// 设置服务客户通道
	s.modifyChannelDataSettings()

	// 从本地磁盘读取路由
	var d1 []byte
	var d2 []byte
	e = s.payRouteMng.LoadAllNodesAndRelationshipFormDisk(s.config.RoutesSourceDataDir, &d1, &d2)
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
	billstore chanpay.DataSourceOfBalanceBill,
	chanset chanpay.DataSourceOfServicerPayChannelSetup,
	signmachine chanpay.DataSourceOfSignatureMachine,
) {
	s.billstore = billstore
	s.chanset = chanset
	s.signmachine = signmachine
}

// 设置数据来源接口
func (s *Servicer) GetLocalServiceNode() (*payroutes.PayRelayNode, error) {
	if s.localServiceNode != nil {
		return s.localServiceNode, nil
	}
	s.customerChgLock.Lock()
	defer s.customerChgLock.Unlock()
	lname := s.config.SelfIdentificationName
	s.localServiceNode = s.payRouteMng.FindNodeByName(lname)
	if s.localServiceNode == nil {
		return nil, fmt.Errorf("Not find PayRelayNode of SelfIdentificationName <%s>", lname)
	}
	return s.localServiceNode, nil
}
