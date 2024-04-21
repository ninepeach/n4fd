package freedom

import (
	"net"

	"github.com/ninepeach/n4fd/common"
	"github.com/ninepeach/n4fd/tunnel"
)

const MaxPacketSize = 1024 * 8

type Conn struct {
	net.Conn
}

func (c *Conn) Read(p []byte) (n int, err error) {
	return c.Conn.Read(p)
}

func (c *Conn) Write(p []byte) (n int, err error) {
	return c.Conn.Write(p)
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return nil
}

type PacketConn struct {
	*net.UDPConn
}

func (c *PacketConn) WriteWithMetadata(p []byte, m *tunnel.Metadata) (int, error) {
	return c.WriteTo(p, m.TargetAddr)
}

func (c *PacketConn) ReadWithMetadata(p []byte) (int, *tunnel.Metadata, error) {
	n, addr, err := c.ReadFrom(p)
	if err != nil {
		return 0, nil, err
	}
	address, err := tunnel.NewAddressFromAddr("udp", addr.String())
	common.Must(err)
	metadata := &tunnel.Metadata{
		TargetAddr: address,
	}
	return n, metadata, nil
}

func (c *PacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		return c.WriteToUDP(p, udpAddr)
	}
	ip, err := addr.(*tunnel.Address).ResolveIP()
	if err != nil {
		return 0, err
	}
	udpAddr := &net.UDPAddr{
		IP:   ip,
		Port: addr.(*tunnel.Address).Port,
	}
	return c.WriteToUDP(p, udpAddr)
}
