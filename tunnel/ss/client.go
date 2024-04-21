package ss

import (
	"context"

	"github.com/ninepeach/n4fd/ss/core"

	"github.com/ninepeach/n4fd/common"
	"github.com/ninepeach/n4fd/config"
	"github.com/ninepeach/n4fd/log"
	"github.com/ninepeach/n4fd/tunnel"
)

type Client struct {
	underlay tunnel.Client
	core.Cipher
}

func (c *Client) DialConn(address *tunnel.Address, tunnel tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(address, &Tunnel{})
	if err != nil {
		return nil, err
	}
	return &Conn{
		aeadConn: c.Cipher.StreamConn(conn),
		Conn:     conn,
	}, nil
}

func (c *Client) DialPacket(tunnel tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (c *Client) Close() error {
	return c.underlay.Close()
}

func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	cipher, err := core.PickCipher(cfg.Shadowsocks.Method, nil, cfg.Shadowsocks.Password)
	if err != nil {
		return nil, common.NewError("invalid shadowsocks cipher").Base(err)
	}
	log.Debug("shadowsocks client created ", cfg.Shadowsocks.Method, cfg.Shadowsocks.Password)
	return &Client{
		underlay: underlay,
		Cipher:   cipher,
	}, nil
}
