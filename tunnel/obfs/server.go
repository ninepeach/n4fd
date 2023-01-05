package obfs

import (
    "net"
    "context"

    log "github.com/ninepeach/go-clog"
    "github.com/ninepeach/n4fd/common"
    "github.com/ninepeach/n4fd/tunnel"
)

type ConnectConn struct {
    net.Conn
    metadata *tunnel.Metadata
}

func (c *ConnectConn) Metadata() *tunnel.Metadata {
    return c.metadata
}

type Server struct {
    underlay     tunnel.Server
    connChan     chan tunnel.Conn
    ctx          context.Context
    cancel       context.CancelFunc
    stage        int8
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
    select {
    case conn := <-s.connChan:
        return conn, nil
    case <-s.ctx.Done():
        return nil, common.NewError("http server closed")
    }
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
    <-s.ctx.Done()
    return nil, common.NewError("http server closed")
}

func (s *Server) Close() error {
    s.cancel()
    return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
    ctx, cancel := context.WithCancel(ctx)

    log.Trace("obfs http server created")

    server := &Server{
        underlay: underlay,
        connChan: make(chan tunnel.Conn, 32),
        ctx:      ctx,
        cancel:   cancel,
    }
    return server, nil
}