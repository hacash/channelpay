package main

import (
	"fmt"
	"github.com/hacash/channelpay/chanpay"
	"github.com/hacash/channelpay/servicer"
	"github.com/hacash/core/sys"
	"os"
	"os/signal"
	"time"
)

func main() {

	// Start the server

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

	// to configure
	svcnf := servicer.NewServicerConfig(hinicnf)
	sev := servicer.NewServicer(svcnf)

	// data source
	localsto, e := chanpay.NewLocalDBImpOfDataSource(svcnf.PaySourceDataDir)
	if e != nil {
		fmt.Println(e.Error())
	} else {
		// set up data sources
		sev.SetDataSource(localsto, localsto, localsto)
		// start-up
		sev.Start()
	}

	// ok
	s := <-c
	fmt.Println("Got signal:", s)

}
