package freedom

import (
	"context"
	"net"

	"github.com/ninepeach/n4fd/common"
	"github.com/ninepeach/n4fd/config"
	"github.com/ninepeach/n4fd/tunnel"
)

type Client struct {
	preferIPv4 bool
	noDelay    bool
	keepAlive  bool
	ctx        context.Context
	cancel     context.CancelFunc
}

func (c *Client) DialConn(addr *tunnel.Address, _ tunnel.Tunnel) (tunnel.Conn, error) {

	network := "tcp"
	if c.preferIPv4 {
		network = "tcp4"
	}
	dialer := new(net.Dialer)
	tcpConn, err := dialer.DialContext(c.ctx, network, addr.String())
	if err != nil {
		return nil, common.NewError("freedom failed to dial " + addr.String()).Base(err)
	}

	tcpConn.(*net.TCPConn).SetKeepAlive(c.keepAlive)
	tcpConn.(*net.TCPConn).SetNoDelay(c.noDelay)
	return &Conn{
		Conn: tcpConn,
	}, nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {

	network := "udp"
	if c.preferIPv4 {
		network = "udp4"
	}
	udpConn, err := net.ListenPacket(network, "")
	if err != nil {
		return nil, common.NewError("freedom failed to listen udp socket").Base(err)
	}
	return &PacketConn{
		UDPConn: udpConn.(*net.UDPConn),
	}, nil
}

func (c *Client) Close() error {
	c.cancel()
	return nil
}

func NewClient(ctx context.Context, _ tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	ctx, cancel := context.WithCancel(ctx)
	return &Client{
		ctx:        ctx,
		cancel:     cancel,
		noDelay:    cfg.TCP.NoDelay,
		keepAlive:  cfg.TCP.KeepAlive,
		preferIPv4: cfg.TCP.PreferIPV4,
	}, nil
}
