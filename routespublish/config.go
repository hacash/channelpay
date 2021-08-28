package routespublish

import (
	"fmt"
	"github.com/hacash/core/sys"
)

/**
 * 配置
 */

type PayRoutesPublishConfig struct {
	WssListenPort int
	DataSourceDir string
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
	fmt.Println(cnf.WssListenPort)
	// data dir
	dtdir := section.Key("data_source_dir").MustString("./hacash_channel_routes_source_data")
	cnf.DataSourceDir = sys.AbsDir(dtdir) // 绝对路径
	// ok
	return cnf
}
