package ss

import (
    "net"
    "fmt"

    "github.com/ninepeach/n4fd/tunnel"
)

type Conn struct {
    aeadConn net.Conn
    tunnel.Conn
}

func (c *Conn) Read(p []byte) (n int, err error) {
    fmt.Println("fucking ss read")
    return c.aeadConn.Read(p)
}

func (c *Conn) Write(p []byte) (n int, err error) {
    fmt.Println("fucking ss write")
    return c.aeadConn.Write(p)
}

func (c *Conn) Close() error {
    c.Conn.Close()
    return c.aeadConn.Close()
}

func (c *Conn) Metadata() *tunnel.Metadata {
    return c.Conn.Metadata()
}
