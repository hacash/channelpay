package main

import (
	"fmt"
	"github.com/hacash/channelpay/servicer"
	"github.com/hacash/channelpay/servicer/datasources"
	"github.com/hacash/core/sys"
	"os"
	"os/signal"
	"time"
)

func main() {

	// 启动服务端

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	target_ini_file := "channelpayservicer.config.ini"

	if len(os.Args) >= 2 {
		target_ini_file = os.Args[1]
	}

	target_ini_file = sys.AbsDir(target_ini_file)

	if target_ini_file != "" {
		fmt.Println("Load ini config file: \"" + target_ini_file + "\" at time:" + time.Now().Format("01/02 15:04:05"))
	}

	hinicnf, _ := sys.LoadInicnf(target_ini_file)

	// 配置
	svcnf := servicer.NewServicerConfig(hinicnf)
	sev := servicer.NewServicer(svcnf)

	// 数据源
	localsto, e := datasources.NewLocalDBImpOfDataSource(svcnf.PaySourceDataDir)
	if e != nil {
		fmt.Println(e.Error())
	} else {
		// 设置数据源
		sev.SetDataSource(localsto, localsto)
		// 启动
		sev.Start()
	}

	// ok
	s := <-c
	fmt.Println("Got signal:", s)

}