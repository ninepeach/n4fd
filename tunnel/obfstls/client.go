package obfstls

import (
	"context"

	"github.com/ninepeach/n4fd/common"
	"github.com/ninepeach/n4fd/log"
	"github.com/ninepeach/n4fd/tunnel"
)

type Client struct {
	underlay tunnel.Client
}

func (c *Client) Close() error {
	return c.underlay.Close()
}

func (c *Client) DialConn(_ *tunnel.Address, overlay tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(nil, &Tunnel{})
	if err != nil {
		return nil, common.NewError("tls failed to dial conn").Base(err)
	}
	return conn, err
}

func (c *Client) DialPacket(tunnel tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {

	log.Debug("obfs-tls client created ")
	return &Client{
		underlay: underlay,
	}, nil
}
