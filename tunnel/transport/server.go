package transport

import (
    "context"
    "bufio"
    "net"
    "net/http"
    "sync"
    "time"

    log "github.com/ninepeach/go-clog"
    "github.com/ninepeach/n4fd/config"
    "github.com/ninepeach/n4fd/common"
    "github.com/ninepeach/n4fd/tunnel"
)

// Server is a server of transport layer
type Server struct {
    tcpListener net.Listener
    connChan    chan tunnel.Conn
    wsChan      chan tunnel.Conn
    httpLock    sync.RWMutex
    nextHTTP    bool
    ctx         context.Context
    cancel      context.CancelFunc
}

func (s *Server) Close() error {
    s.cancel()
    return s.tcpListener.Close()
}

func (s *Server) acceptLoop() {
    for {
        tcpConn, err := s.tcpListener.Accept()
        if err != nil {
            select {
            case <-s.ctx.Done():
            default:
                log.Error(common.NewError("transport accept error").Base(err).Error())
                time.Sleep(time.Millisecond * 100)
            }
            return
        }

        go func(tcpConn net.Conn) {
            log.Info("tcp connection from %s", tcpConn.RemoteAddr())
            s.httpLock.RLock()
            if s.nextHTTP { // plaintext mode enabled
                s.httpLock.RUnlock()
                // we use real http header parser to mimic a real http server
                rewindConn := common.NewRewindConn(tcpConn)
                rewindConn.SetBufferSize(512)
                defer rewindConn.StopBuffering()

                r := bufio.NewReader(rewindConn)
                httpReq, err := http.ReadRequest(r)
                rewindConn.Rewind()
                rewindConn.StopBuffering()
                if err != nil {
                    // this is not a http request, pass it to trojan protocol layer for further inspection
                    s.connChan <- &Conn{
                        Conn: rewindConn,
                    }
                } else {
                    // this is a http request, pass it to websocket protocol layer
                    log.Debug("plaintext http request: %v", httpReq)
                    s.wsChan <- &Conn{
                        Conn: rewindConn,
                    }
                }
            } else {
                s.httpLock.RUnlock()
                s.connChan <- &Conn{
                    Conn: tcpConn,
                }
            }
        }(tcpConn)
    }
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
    // TODO fix import cycle
    if overlay != nil && (overlay.Name() == "WEBSOCKET" || overlay.Name() == "HTTP") {
        s.httpLock.Lock()
        s.nextHTTP = true
        s.httpLock.Unlock()
        select {
        case conn := <-s.wsChan:
            return conn, nil
        case <-s.ctx.Done():
            return nil, common.NewError("transport server closed")
        }
    }
    select {
    case conn := <-s.connChan:
        return conn, nil
    case <-s.ctx.Done():
        return nil, common.NewError("transport server closed")
    }
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
    panic("not supported")
}

// NewServer creates a transport layer server
func NewServer(ctx context.Context, _ tunnel.Server) (*Server, error) {
    cfg := config.FromContext(ctx, Name).(*Config)
    listenAddress := tunnel.NewAddressFromHostPort("tcp", cfg.LocalHost, cfg.LocalPort)

    log.Info("listen on %s", listenAddress.String())
    tcpListener, err := net.Listen("tcp", listenAddress.String())
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithCancel(ctx)
    server := &Server{
        tcpListener: tcpListener,
        ctx:         ctx,
        cancel:      cancel,
        connChan:    make(chan tunnel.Conn, 32),
        wsChan:      make(chan tunnel.Conn, 32),
    }
    go server.acceptLoop()
    return server, nil
}
