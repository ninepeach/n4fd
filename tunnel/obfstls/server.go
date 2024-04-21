package obfstls

import (
	"context"
	"errors"

	"github.com/ninepeach/n4fd/common"
	"github.com/ninepeach/n4fd/log"
	"github.com/ninepeach/n4fd/tunnel"
	"github.com/ninepeach/n4fd/tunnel/transport"
)

type Server struct {
	underlay tunnel.Server
	connChan chan tunnel.Conn
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Server) Name() string {
	return Name
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.underlay.AcceptConn(&Tunnel{})

		if err != nil { // Closing
			log.Error(common.NewError("obfs-tls got failed to accept conn").Base(err))
			select {
			case <-s.ctx.Done():
				return
			default:
			}
			return
		}
		go func(conn tunnel.Conn) {

			newConn := &Conn{
				Conn:       conn,
				handshaked: make(chan struct{}),
			}

			s.connChan <- &transport.Conn{
				Conn: newConn,
			}

		}(conn)
	}
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, errors.New("objs-tls server closed")
	}
}

func (s *Server) AcceptPacket(t tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (s *Server) Close() error {
	return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	//cfg := config.FromContext(ctx, Name).(*Config)

	log.Debug("obfs-tls server created ")

	ctx, cancel := context.WithCancel(ctx)
	s := &Server{
		underlay: underlay,
		connChan: make(chan tunnel.Conn, 32),
		ctx:      ctx,
		cancel:   cancel,
	}
	go s.acceptLoop()
	return s, nil
}
