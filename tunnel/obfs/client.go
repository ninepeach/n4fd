package obfs

import (
    "context"

    log "github.com/ninepeach/go-clog"
    "github.com/ninepeach/n4fd/tunnel"
)

type Client struct {
    underlay tunnel.Client
}

func (c *Client) DialConn(address *tunnel.Address, tunnel tunnel.Tunnel) (tunnel.Conn, error) {
    conn, err := c.underlay.DialConn(address, &Tunnel{})
    if err != nil {
        return nil, err
    }
    return &Conn{
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
    log.Debug("obfs client created")
    return &Client{
        underlay: underlay,
    }, nil
}