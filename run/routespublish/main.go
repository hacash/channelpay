package main

import (
	"fmt"
	"github.com/hacash/channelpay/routespublish"
	"github.com/hacash/core/sys"
	"os"
	"os/signal"
	"time"
)

func main() {

	// 启动服务端

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	target_ini_file := "channelpayroutespublish.config.ini"

	if len(os.Args) >= 2 {
		target_ini_file = os.Args[1]
	}

	target_ini_file = sys.AbsDir(target_ini_file)

	if target_ini_file != "" {
		fmt.Println("Load ini config file: \"" + target_ini_file + "\" at time:" + time.Now().Format("01/02 15:04:05"))
	}

	hinicnf, _ := sys.LoadInicnf(target_ini_file)

	// 配置
	svcnf := routespublish.NewPayRoutesPublishConfig(hinicnf)
	sev := routespublish.NewPayRoutesPublish(svcnf)

	// 启动
	sev.Start()

	// ok
	s := <-c
	fmt.Println("Got signal:", s)

}
