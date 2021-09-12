package servicer

import (
	"github.com/hacash/core/sys"
	"os"
)

type ServicerConfig struct {
	DebugTest bool

	WssListenPort int

	PaySourceDataDir    string // 支付数据储存
	RoutesSourceDataDir string // 路由数据储存

	SelfIdentificationName string // 本机服务商识别名称
	LoadRoutesUrl          string // 下发路由数据

	FullNodeRpcUrl string // Hacash 全节点数据接口地址

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
	// ok
	return cnf
}
