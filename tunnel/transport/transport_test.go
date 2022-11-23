package transport

import (
    "context"
    "net"
    "sync"
    "testing"

    "bytes"
    "crypto/rand"

    "github.com/ninepeach/n4fd/config"
    "github.com/ninepeach/n4fd/common"
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

func TestTransport(t *testing.T) {
    serverCfg := &Config{
        LocalHost:  "127.0.0.1",
        LocalPort:  common.PickPort("tcp", "127.0.0.1"),
        RemoteHost: "127.0.0.1",
        RemotePort: common.PickPort("tcp", "127.0.0.1"),
    }
    clientCfg := &Config{
        LocalHost:  "127.0.0.1",
        LocalPort:  common.PickPort("tcp", "127.0.0.1"),
        RemoteHost: serverCfg.LocalHost,
        RemotePort: serverCfg.LocalPort,
    }
    sctx := config.WithConfig(context.Background(), Name, serverCfg)
    cctx := config.WithConfig(context.Background(), Name, clientCfg)

    s, err := NewServer(sctx, nil)
    common.Must(err)
    c, err := NewClient(cctx, nil)
    common.Must(err)


    wg := sync.WaitGroup{}
    wg.Add(1)
    var conn1, conn2 net.Conn
    go func() {
        conn2, err = s.AcceptConn(nil)
        common.Must(err)
        wg.Done()
    }()
    conn1, err = c.DialConn(nil, nil)
    common.Must(err)

    common.Must2(conn1.Write([]byte("12345678\r\n")))
    wg.Wait()
    buf := [10]byte{}
    conn2.Read(buf[:])

    if !CheckConn(conn1, conn2) {
        t.Fail()
    }
    s.Close()
    c.Close()
    
}