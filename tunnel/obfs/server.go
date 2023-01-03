package obfs

import (
    "net"
    "strings"
    "context"
    "reflect"

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

func (s *Server) acceptLoop() {
    for {
        conn, err := s.underlay.AcceptConn(&Tunnel{})
        //czw debug
        log.Debug("accept conn underlay:%v", reflect.TypeOf(s.underlay))

        if err != nil {
            select {
            case <-s.ctx.Done():
                log.Debug(common.NewError("http closed").Error())
                return
            default:
                log.Error(common.NewError("http failed to accept connection").Base(err).Error())
                continue
            }
        }

        log.Info("acceptLoop: -> go func")

        go func(conn net.Conn) {

            rewindConn := common.NewRewindConn(conn)
            rewindConn.SetBufferSize(512)
            defer rewindConn.StopBuffering()

            // read
            buf := make([]byte, 171)
            n, err := rewindConn.Read(buf)
            if err != nil {
                // handle error
            }
            // find end of header
            headerEnd := strings.Index(string(buf[:n]), "\r\n\r\n")
            if headerEnd == -1 {
                // handle error: end of header not found
            }
            log.Info("header end %d", headerEnd)

            rewindConn.Rewind()
            rewindConn.StopBuffering()


        }(conn)
    }
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
    server := &Server{
        underlay: underlay,
        connChan: make(chan tunnel.Conn, 32),
        ctx:      ctx,
        cancel:   cancel,
        stage:    0,
    }
    go server.acceptLoop()
    return server, nil
}
