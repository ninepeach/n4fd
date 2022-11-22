package utils

import (
    "fmt"
    "net/http"
    "time"

    "golang.org/x/net/websocket"

    log "github.com/ninepeach/go-clog"
    "github.com/ninepeach/n4fd/common"
)

func runHelloHTTPServer(addr string) {

	httpHello := func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("HelloWorld"))
	}

	wsConfig, err := websocket.NewConfig("wss://127.0.0.1/websocket", "https://127.0.0.1")
	common.Must(err)
	wsServer := websocket.Server{
		Config: *wsConfig,
		Handler: func(conn *websocket.Conn) {
			conn.Write([]byte("HelloWorld"))
		},
		Handshake: func(wsConfig *websocket.Config, httpRequest *http.Request) error {
			log.Trace("websocket url %s %s %s", httpRequest.URL, "origin", httpRequest.Header.Get("Origin"))
			return nil
		},
	}
	mux := &http.ServeMux{}
	mux.HandleFunc("/", httpHello)
	mux.HandleFunc("/websocket", wsServer.ServeHTTP)

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}
	go server.ListenAndServe()
	time.Sleep(time.Second * 1) // wait for http server
	fmt.Println("http test server listening on", addr)
}

