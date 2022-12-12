package main

import (
    "fmt"
    "net"
    "context"
    log "github.com/ninepeach/go-clog"
    "github.com/ninepeach/n4fd/config"
    "github.com/ninepeach/n4fd/common"
    "github.com/ninepeach/n4fd/tunnel/transport"
    "github.com/ninepeach/n4fd/tunnel/ss"
)

const Name = "SS"

func acceptLoop(s *ss.Server) {
    for {
        conn, err := s.AcceptConn(nil)
        if err != nil {
            log.InfoTo(log.DefaultConsoleName, "shadowsocks failed to accept connection %s!", err.Error())
            break
        }

        go func(conn net.Conn) {
            fmt.Println(conn)
        }(conn)
    }
}

func main() {
    fmt.Println("fucking cao")

    transportConfig := &transport.Config{
        LocalHost:  "0.0.0.0",
        LocalPort:  12022,
        RemoteHost: "127.0.0.1",
        RemotePort: common.PickPort("tcp", "127.0.0.1"),
    }

    ctx := config.WithConfig(context.Background(), transport.Name, transportConfig)
    tcpServer, err := transport.NewServer(ctx, nil)
    common.Must(err)

    cfg := &ss.Config{
        RemoteHost: "127.0.0.1",
        RemotePort: common.PickPort("tcp", "127.0.0.1"),
        Shadowsocks: ss.ShadowsocksConfig{
            Enabled:  true,
            Method:   "AES-128-GCM",
            Password: "password",
        },
    }
    ctx = config.WithConfig(ctx, Name, cfg)
    s, err := ss.NewServer(ctx, tcpServer)
    fmt.Println(s, err)
    log.InfoTo(log.DefaultConsoleName, "Hello %s!", "World")
    acceptLoop(s)
}
