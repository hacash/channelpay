package servicer

import (
	"github.com/hacash/core/sys"
	"os"
)

type ServicerConfig struct {
	DebugTest bool

	WssListenPort int

	PaySourceDataDir    string // Payment data storage
	RoutesSourceDataDir string // Routing data storage

	SelfIdentificationName string // Local service provider identification name
	LoadRoutesUrl          string // Send routing data

	FullNodeRpcUrl string // Hacash all node data interface address

	// Data modification
	ServiceCustomerChannelsAdd    string // List of customer service channels to add
	ServiceCustomerChannelsCancel string // Cancel service

	// Node settlement channel profile
	RelaySettlementChannelsJsonFile string

	// Signer private key list
	SignatureMachinePrivateKeySetupList string
}

func NewEmptyServicerConfig() *ServicerConfig {
	cnf := &ServicerConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewServicerConfig(cnffile *sys.Inicnf) *ServicerConfig {
	cnf := NewEmptyServicerConfig()
	section := cnffile.Section("")

	// debug
	cnf.DebugTest = section.Key("DebugTest").MustBool(false)

	// port
	cnf.WssListenPort = section.Key("listen_port").MustInt(3351)

	// data dir
	dir1 := section.Key("pay_source_data_dir").MustString("./hacash_channel_pay_source_data")
	cnf.PaySourceDataDir = sys.AbsDir(dir1)
	os.MkdirAll(cnf.PaySourceDataDir, 0777)

	dir2 := section.Key("routes_source_data_dir").MustString("./hacash_channel_routes_source_data")
	cnf.RoutesSourceDataDir = sys.AbsDir(dir2)
	os.MkdirAll(cnf.RoutesSourceDataDir, 0777)

	// service name
	cnf.SelfIdentificationName = section.Key("pay_servicer_identification_name").MustString("HACorg")

	cnf.LoadRoutesUrl = section.Key("load_routes_data_url").MustString("wss://channelroutes.hacash.org/routesdata/distribute")

	cnf.FullNodeRpcUrl = section.Key("full_node_rpc_url").MustString("http://127.0.0.1:3381")

	// Data node
	section2 := cnffile.Section("channel")
	// List of customer service channels to add
	cnf.ServiceCustomerChannelsAdd = section2.Key("service_customer_channels_add").MustString("")
	// Cancel service
	cnf.ServiceCustomerChannelsCancel = section2.Key("service_customer_channels_cancel").MustString("")
	// Inter node settlement channel
	cnf.RelaySettlementChannelsJsonFile = section2.Key("relay_settlement_channels_json_file").MustString("./hacash_relay_settlement_channels_json_file.json")

	// password
	section3 := cnffile.Section("password")
	// Signer private key
	cnf.SignatureMachinePrivateKeySetupList = section3.Key("signature_machine_password_setup_list").MustString("")
	// ok
	return cnf
}
