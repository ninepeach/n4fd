package transport

import (
	"context"

	"github.com/ninepeach/n4fd/common"
	"github.com/ninepeach/n4fd/config"
	"github.com/ninepeach/n4fd/tunnel"
	"github.com/ninepeach/n4fd/tunnel/freedom"
)

// Client implements tunnel.Client
type Client struct {
	serverAddress *tunnel.Address
	ctx           context.Context
	cancel        context.CancelFunc
	direct        *freedom.Client
}

func (c *Client) Close() error {
	c.cancel()
	return nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

// DialConn implements tunnel.Client. It will ignore the params and directly dial to the remote server
func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.direct.DialConn(c.serverAddress, nil)
	if err != nil {
		return nil, common.NewError("transport failed to connect to remote server").Base(err)
	}
	return &Conn{
		Conn: conn,
	}, nil
}

// NewClient creates a transport layer client
func NewClient(ctx context.Context, _ tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)

	serverAddress := tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort)

	direct, err := freedom.NewClient(ctx, nil)
	common.Must(err)
	ctx, cancel := context.WithCancel(ctx)
	client := &Client{
		serverAddress: serverAddress,
		ctx:           ctx,
		cancel:        cancel,
		direct:        direct,
	}
	return client, nil
}
