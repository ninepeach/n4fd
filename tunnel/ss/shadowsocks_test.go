package ss

import (
    "context"
    "net"
    "sync"
    "bytes"
    "crypto/rand"
    "testing"

    "github.com/ninepeach/n4fd/common"
    "github.com/ninepeach/n4fd/config"
    "github.com/ninepeach/n4fd/tunnel/transport"
)

func CheckConn(a net.Conn, b net.Conn) bool {
    payload1 := [1024]byte{}
    payload2 := [1024]byte{}
    rand.Reader.Read(payload1[:])
    rand.Reader.Read(payload2[:])

    result1 := [1024]byte{}
    result2 := [1024]byte{}
    wg := sync.WaitGroup{}
    wg.Add(2)
    go func() {
        a.Write(payload1[:])
        a.Read(result2[:])
        wg.Done()
    }()
    go func() {
        b.Read(result1[:])
        b.Write(payload2[:])
        wg.Done()
    }()
    wg.Wait()
    if !bytes.Equal(payload1[:], result1[:]) || !bytes.Equal(payload2[:], result2[:]) {
        return false
    }
    return true
}

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
    if !CheckConn(conn1, conn2) {
        t.Fail()
    }


    conn1.Close()
    conn2.Close()
    c.Close()
    s.Close()
}
