package obfs

import (
    "fmt"
    "github.com/ninepeach/n4fd/tunnel"
)

type Conn struct {
    tunnel.Conn
}

func (c *Conn) Read(p []byte) (n int, err error) {
    return c.Conn.Read(p)
}

func (c *Conn) Write(p []byte) (n int, err error) {
    fmt.Println("fucking cao %v", p)
    c.Conn.Write(p)
    return c.Conn.Write(p)
}

func (c *Conn) Close() error {
    c.Conn.Close()
    return c.Conn.Close()
}

func (c *Conn) Metadata() *tunnel.Metadata {
    return c.Conn.Metadata()
}
