package servicer

import (
	"fmt"
	"github.com/hacash/node/websocket"
	"log"
	"net/http"
	"os"
	"strconv"
)

func (s *Servicer) startListen() {

	mux := http.NewServeMux()

	// 处理顾客连接
	mux.Handle("/customer/connect", websocket.Handler(s.connectHandler))

	// 设置监听的端口
	portstr := strconv.Itoa(s.config.WssListenPort)

	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[Listen] Servicer listen on port: " + portstr)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe Error: ", err)
		os.Exit(0)
	} else {
		fmt.Println("[Listen] Successfully listen on port: " + portstr)
	}

}
