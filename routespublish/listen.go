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
		//w.Write([]byte(`{"server":"HacashChannelPayRoutesPublish"}`))
		w.Write([]byte(GetHomePageHtml()))
	})

	// Websocket downloading distribution channel routing data
	mux.Handle("/routesdata/distribute", websocket.Handler(p.connectHandler))

	// Channel chain user login resolution
	mux.HandleFunc("/customer/login_resolution", p.customerLoginResolution)
	mux.HandleFunc("/customer/hdns_analyze", p.customerAnalyzeHDNS)

	// Set listening port
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

// Processing messages
func (p *PayRoutesPublish) connectHandler(ws *websocket.Conn) {

	for {
		// Read message
		msgobj, _, err := protocol.ReceiveMsg(ws)
		if err != nil {
			break
		}
		// Message must request update or request all data
		mt := msgobj.Type()
		if mt == protocol.MsgTypePayRouteRequestServiceNodes {
			// Send all service node information
			p.SendAllNodesData(ws)

		} else if mt == protocol.MsgTypePayRouteRequestNodeRelationship {
			// Send all service node information
			p.SendAllGraphData(ws)

		} else if mt == protocol.MsgTypePayRouteRequestUpdates {
			// Send all service node information
			upmg := msgobj.(*protocol.MsgPayRouteRequestUpdates)
			p.SendTargetUpdateData(ws, uint32(upmg.QueryPageNum))

		} else if mt == protocol.MsgTypePayRouteEndClose {
			// Complete and close
			break

		} else {
			// Unsupported message type, disconnect directly
			break
		}

	}

	// Disconnect
	ws.Close()
}
