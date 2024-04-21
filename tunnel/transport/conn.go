package transport

import (
	"net"

	"github.com/ninepeach/n4fd/tunnel"
)

type Conn struct {
	net.Conn
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return nil
}
