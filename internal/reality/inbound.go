package reality

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	xtls "github.com/xtls/reality"
)

// InboundConfig holds listener + REALITY options.
type InboundConfig struct {
	Listen   string
	PrivKey  []byte
	SNIAllow map[string]bool   // nil => allow any
	ShortIds map[[8]byte]bool  // allow zero if present
	Dest     string            // mirror dest, e.g. "www.cloudflare.com:443"
	Debug    bool
	Timeout  time.Duration
}

// NextHandler handles a post-REALITY-OK connection (VLESS parser).
type NextHandler interface {
	HandleConn(net.Conn) error
}

type Inbound struct {
	cfg  InboundConfig
	next NextHandler
	rcfg *xtls.Config
}

func NewInbound(cfg InboundConfig, next NextHandler) *Inbound {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	dialer := &net.Dialer{Timeout: 8 * time.Second, KeepAlive: 30 * time.Second}
	rc := &xtls.Config{
		DialContext:            dialer.DialContext,
		Show:                   cfg.Debug,
		Type:                   "tcp",
		Dest:                   cfg.Dest,
		PrivateKey:             cfg.PrivKey,
		ServerNames:            cfg.SNIAllow,
		ShortIds:               cfg.ShortIds,
		MinVersion:             tls.VersionTLS13,
		NextProtos:             nil,
		SessionTicketsDisabled: true,
		MaxTimeDiff:            0,
	}
	return &Inbound{cfg: cfg, next: next, rcfg: rc}
}

func (in *Inbound) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", in.cfg.Listen)
	if err != nil {
		return fmt.Errorf("listen %s: %w", in.cfg.Listen, err)
	}
	log.Printf("REALITY inbound listening addr=%s dest=%s sni_allow=%v shortids=%d",
		in.cfg.Listen, in.cfg.Dest, in.cfg.SNIAllow, len(in.cfg.ShortIds))

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		c, err := ln.Accept()
		if err != nil {
			return err
		}
		go in.serveConn(c)
	}
}

type closeWriter interface{ CloseWrite() error }

type cwConn struct{ net.Conn }

func (c *cwConn) CloseWrite() error {
	if x, ok := c.Conn.(closeWriter); ok {
		return x.CloseWrite()
	}
	return nil
}

func (in *Inbound) serveConn(c net.Conn) {
	defer c.Close()
	remote := c.RemoteAddr().String()

	br := bufio.NewReader(c)
	if b, _ := br.Peek(1); len(b) == 1 && in.cfg.Debug {
		log.Printf("DEBUG TCP prelude remote=%s b0=0x%02x", remote, b[0])
	}
	rw := &cwConn{Conn: &peekBack{Conn: c, r: br}}

	ctx, cancel := context.WithTimeout(context.Background(), in.cfg.Timeout)
	defer cancel()

	rconn, err := safeServer(ctx, rw, in.rcfg)
	if err != nil {
		log.Printf("REALITY aborted remote=%s err=%v", remote, err)
		return
	}

	if err := in.next.HandleConn(rconn); err != nil {
		log.Printf("VLESS handler error remote=%s err=%v", remote, err)
	}
}

// safeServer wraps reality.Server to avoid panics
func safeServer(ctx context.Context, c net.Conn, rc *xtls.Config) (_ *xtls.Conn, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("reality handshake aborted")
		}
	}()
	return xtls.Server(ctx, c, rc)
}

// peekBack preserves buffered bytes
type peekBack struct {
	net.Conn
	r *bufio.Reader
}

func (p *peekBack) Read(b []byte) (int, error) {
	return p.r.Read(b)
}
