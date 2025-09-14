package direct

import (
	"net"
	"time"
)

var DefaultDialer = &net.Dialer{Timeout: 15 * time.Second}

func Dial(network, addr string) (net.Conn, error) {
	return DefaultDialer.Dial(network, addr)
}
