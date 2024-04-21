package ss

import (
	"github.com/ninepeach/n4fd/tunnel"
	"net"
)

type Conn struct {
	net.Conn
	aeadConn net.Conn
	metadata *tunnel.Metadata
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return c.aeadConn.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return c.aeadConn.Write(b)
}

func (c *Conn) Close() error {
	c.Conn.Close()
	return c.aeadConn.Close()
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return c.metadata
}
