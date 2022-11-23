package transport

import (
    "net"
    "context"
    "fmt"

    "github.com/ninepeach/n4fd/config"
    "github.com/ninepeach/n4fd/common"
    "github.com/ninepeach/n4fd/tunnel"
)

// Client implements tunnel.Client
type Client struct {
    ctx           context.Context
    cancel        context.CancelFunc

    serverAddress *tunnel.Address
    noDelay       bool
    keepAlive     bool
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

    network := "tcp"

    dialer := new(net.Dialer)
    fmt.Println(network)
    tcpConn, err := dialer.DialContext(c.ctx, network, c.serverAddress.String())
    if err != nil {
        return nil, common.NewError("transport failed to dial " + c.serverAddress.String()).Base(err)
    }

    tcpConn.(*net.TCPConn).SetKeepAlive(c.keepAlive)
    tcpConn.(*net.TCPConn).SetNoDelay(c.noDelay)
    return &Conn{
        Conn: tcpConn,
    }, nil

}

// NewClient creates a transport layer client
func NewClient(ctx context.Context, _ tunnel.Client) (*Client, error) {
    cfg := config.FromContext(ctx, Name).(*Config)

    serverAddress := tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort)

    ctx, cancel := context.WithCancel(ctx)
    client := &Client{
        ctx:          ctx,
        cancel:       cancel,
        serverAddress: serverAddress,
        noDelay:      cfg.TCP.NoDelay,
        keepAlive:    cfg.TCP.KeepAlive,
    }
    return client, nil
}
