package ss

import (
    "fmt"
    "context"
    "net"

    "github.com/ninepeach/go-libss/core"

    log "github.com/ninepeach/go-clog"

    "github.com/ninepeach/n4fd/common"
    "github.com/ninepeach/n4fd/config"
    "github.com/ninepeach/n4fd/redirector"
    "github.com/ninepeach/n4fd/tunnel"
)

type Server struct {
    core.Cipher
    *redirector.Redirector
    underlay  tunnel.Server
    redirAddr net.Addr
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
    conn, err := s.underlay.AcceptConn(&Tunnel{})
    if err != nil {
        return nil, common.NewError("shadowsocks failed to accept connection from underlying tunnel").Base(err)
    }
    rewindConn := common.NewRewindConn(conn)
    rewindConn.SetBufferSize(1024)
    defer rewindConn.StopBuffering()

    // try to read something from this connection
    buf := [1024]byte{}
    testConn := s.Cipher.StreamConn(rewindConn)
    if _, err := testConn.Read(buf[:]); err != nil {
        // we are under attack
        log.Error(common.NewError("shadowsocks failed to decrypt").Base(err).Error())
        rewindConn.Rewind()
        rewindConn.StopBuffering()
        s.Redirect(&redirector.Redirection{
            RedirectTo:  s.redirAddr,
            InboundConn: rewindConn,
        })
        return nil, common.NewError("invalid aead payload")
    }
    rewindConn.Rewind()
    rewindConn.StopBuffering()

    return &Conn{
        aeadConn: s.Cipher.StreamConn(rewindConn),
        Conn:     conn,
    }, nil
}

func (s *Server) AcceptPacket(t tunnel.Tunnel) (tunnel.PacketConn, error) {
    panic("not supported")
}

func (s *Server) Close() error {
    return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
    cfg := config.FromContext(ctx, Name).(*Config)
    // czw debug
    fmt.Println(cfg)
    cipher, err := core.PickCipher(cfg.Shadowsocks.Method, nil, cfg.Shadowsocks.Password)
    if err != nil {
        return nil, common.NewError("invalid shadowsocks cipher").Base(err)
    }
    log.Trace("shadowsocks server created")
    return &Server{
        underlay:   underlay,
        Cipher:     cipher,
        Redirector: redirector.NewRedirector(ctx),
        redirAddr:  tunnel.NewAddressFromHostPort("tcp", "8.8.8.8", 88),
    }, nil
}
