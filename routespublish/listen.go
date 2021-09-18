package routespublish

import (
	"fmt"
	"github.com/hacash/channelpay/protocol"
	"github.com/hacash/node/websocket"
	"log"
	"net/http"
	"os"
	"strconv"
)

/**
 * wss 监听端口
 */
func (p *PayRoutesPublish) listen(port int) {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"server":"HacashChannelPayRoutesPublish"}`))
	})

	// websocket 下载分发通道路由数据
	mux.Handle("/routesdata/distribute", websocket.Handler(p.connectHandler))

	// 通道链用户登录解析
	mux.HandleFunc("/customer/login_resolution", p.customerLoginResolution)

	// 设置监听的端口
	portstr := strconv.Itoa(port)

	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[Listen] PayRoutesPublish listen on port: " + portstr)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe Error: ", err)
		os.Exit(0)
	} else {
		fmt.Println("[Listen] Successfully listen on port: " + portstr)
	}

}

// 处理消息
func (p *PayRoutesPublish) connectHandler(ws *websocket.Conn) {

	for {
		// 读取消息
		msgobj, _, err := protocol.ReceiveMsg(ws)
		if err != nil {
			break
		}
		// 消息必须请求更新或请求全部数据
		mt := msgobj.Type()
		if mt == protocol.MsgTypePayRouteRequestServiceNodes {
			// 发送全部服务节点信息
			p.SendAllNodesData(ws)

		} else if mt == protocol.MsgTypePayRouteRequestNodeRelationship {
			// 发送全部服务节点信息
			p.SendAllGraphData(ws)

		} else if mt == protocol.MsgTypePayRouteRequestUpdates {
			// 发送全部服务节点信息
			upmg := msgobj.(*protocol.MsgPayRouteRequestUpdates)
			p.SendTargetUpdateData(ws, uint32(upmg.QueryPageNum))

		} else if mt == protocol.MsgTypePayRouteEndClose {
			// 完成并关闭
			break

		} else {
			// 不支持的消息类型，直接断开
			break
		}

	}

	// 断开连接
	ws.Close()
}
