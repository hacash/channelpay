package routespublish

import (
	"github.com/hacash/core/sys"
)

/**
 * 配置
 */

type PayRoutesPublishConfig struct {
	WssListenPort int
	DataSourceDir string

	FullNodeRpcURL string // All node RPC address
}

func NewEmptyPayRoutesPublishConfig() *PayRoutesPublishConfig {
	cnf := &PayRoutesPublishConfig{}
	return cnf
}

//////////////////////////////////////////////////

func NewPayRoutesPublishConfig(cnffile *sys.Inicnf) *PayRoutesPublishConfig {
	cnf := NewEmptyPayRoutesPublishConfig()
	section := cnffile.Section("")
	// port
	cnf.WssListenPort = section.Key("listen_port").MustInt(3350)
	// data dir
	dtdir := section.Key("data_source_dir").MustString("./hacash_channel_routes_source_data")
	cnf.DataSourceDir = sys.AbsDir(dtdir) // Absolute path
	// rpc
	cnf.FullNodeRpcURL = section.Key("full_node_rpc_url").MustString("http://127.0.0.1:38082")

	// ok
	return cnf
}
