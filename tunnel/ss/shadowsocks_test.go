package ss

import (
    "context"
    "net"
    "sync"
    "testing"

    "github.com/ninepeach/n4fd/common"
    "github.com/ninepeach/n4fd/config"
    "github.com/ninepeach/n4fd/tunnel/transport"
    "github.com/ninepeach/n4fd/test/util"
)

func TestShadowsocks(t *testing.T) {

    port := common.PickPort("tcp", "127.0.0.1")
    transportConfig := &transport.Config{
        LocalHost:  "127.0.0.1",
        LocalPort:  port,
        RemoteHost: "127.0.0.1",
        RemotePort: port,
    }

    ctx := config.WithConfig(context.Background(), transport.Name, transportConfig)
    tcpClient, err := transport.NewClient(ctx, nil)
    common.Must(err)
    tcpServer, err := transport.NewServer(ctx, nil)
    common.Must(err)

    cfg := &Config{
        RemoteHost: "127.0.0.1",
        RemotePort: 1341,
        Shadowsocks: ShadowsocksConfig{
            Enabled:  true,
            Method:   "AES-128-GCM",
            Password: "password",
        },
    }
    ctx = config.WithConfig(ctx, Name, cfg)
    c, err := NewClient(ctx, tcpClient)
    common.Must(err)
    s, err := NewServer(ctx, tcpServer)
    common.Must(err)

    wg := sync.WaitGroup{}
    wg.Add(2)
    var conn1, conn2 net.Conn
    go func() {
        var err error
        conn1, err = c.DialConn(nil, nil)
        common.Must(err)
        conn1.Write([]byte("12345678\r\n"))
        wg.Done()
    }()
    go func() {
        var err error
        conn2, err = s.AcceptConn(nil)
        common.Must(err)
        buf := [12]byte{}
        conn2.Read(buf[:])
        wg.Done()
    }()
    wg.Wait()
    if !util.CheckConn(conn1, conn2) {
        t.Fail()
    }


    conn1.Close()
    conn2.Close()
    c.Close()
    s.Close()
}
