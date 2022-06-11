package servicer

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/payroutes"
	"sync"
)

type Servicer struct {
	config *ServicerConfig

	// data interface 
	billstore   chanpay.DataSourceOfBalanceBill             // Bill storage
	chanset     chanpay.DataSourceOfServicerPayChannelSetup // Channel configuration
	signmachine chanpay.DataSourceOfSignatureMachine        // Signing machine

	// Customer connection pool
	customerChgLock sync.RWMutex
	customers       map[uint64]*chanpay.Customer // client

	// Settlement channel connection pool
	settlenoderChgLock sync.RWMutex
	settlenoder        map[string][]*chanpay.RelayPaySettleNoder // Settlement channel

	// Routing manager
	payRouteMng      *payroutes.RoutingManager
	localServiceNode *payroutes.PayRelayNode // Local service node
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

// start-up
func (s *Servicer) Start() {

	var e error

	// Initialize settlement channel
	s.setupRelaySettlementChannelDataSettings()

	// Initialize settlement channel
	s.setupPasswordSettings()

	// Set up service customer channel
	s.modifyChannelDataSettings()

	// Read route from local disk
	var d1 []byte
	var d2 []byte
	e = s.payRouteMng.LoadAllNodesAndRelationshipFormDisk(s.config.RoutesSourceDataDir, &d1, &d2)
	if e != nil {
		fmt.Println(e)
	}

	// Load route from remote
	s.checkInitLoadRoutes()

	go s.loop()

	go s.startListen()

}

// Set data source interface
func (s *Servicer) SetDataSource(
	billstore chanpay.DataSourceOfBalanceBill,
	chanset chanpay.DataSourceOfServicerPayChannelSetup,
	signmachine chanpay.DataSourceOfSignatureMachine,
) {
	s.billstore = billstore
	s.chanset = chanset
	s.signmachine = signmachine
}

// Set data source interface
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
