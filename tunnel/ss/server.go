package ss

import (
	"context"
	"errors"
	"fmt"
	"github.com/ninepeach/n4fd/ss/core"
	"github.com/ninepeach/n4fd/ss/socks"

	"github.com/ninepeach/n4fd/common"
	"github.com/ninepeach/n4fd/config"
	"github.com/ninepeach/n4fd/log"
	"github.com/ninepeach/n4fd/tunnel"
)

type Server struct {
	core.Cipher
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
			log.Error(common.NewError("ss failed to accept conn").Base(err))
			select {
			case <-s.ctx.Done():
				return
			default:
			}
			continue
		}
		go func(conn tunnel.Conn) {

			aConn := s.Cipher.StreamConn(conn)
			newConn := &Conn{
				Conn:     conn,
				aeadConn: aConn,
			}

			addr, err := socks.ReadAddr(aConn)
			if err != nil {
				log.Error(common.NewError(fmt.Sprint("failed to get target address from ", conn.RemoteAddr())).Base(err))
				newConn.Close()
				return
			}

			target, err := tunnel.NewAddressFromAddr("tcp", addr.String())
			remote, err := tunnel.NewAddressFromAddr("tcp", conn.RemoteAddr().String())

			newConn.metadata = &tunnel.Metadata{
				TargetAddr: target,
				RemoteAddr: remote,
			}
			s.connChan <- newConn
		}(conn)
	}
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, errors.New("ss server closed")
	}
}

func (s *Server) AcceptPacket(t tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	cipher, err := core.PickCipher(cfg.Shadowsocks.Method, nil, cfg.Shadowsocks.Password)
	if err != nil {
		return nil, common.NewError("invalid shadowsocks cipher").Base(err)
	}
	log.Debug("shadowsocks server created ", cfg.Shadowsocks.Method, cfg.Shadowsocks.Password)

	ctx, cancel := context.WithCancel(ctx)
	s := &Server{
		underlay: underlay,
		Cipher:   cipher,
		connChan: make(chan tunnel.Conn, 32),
		ctx:      ctx,
		cancel:   cancel,
	}
	go s.acceptLoop()
	return s, nil
}
