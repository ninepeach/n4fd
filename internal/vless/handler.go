package vless

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// Handler performs minimal VLESS TCP processing:
// - parse version(=0), UUID(16B), command, address, port
// - whitelist UUID, then direct-dial target and splice.
type Handler struct {
	Allow     map[[16]byte]bool
	Dialer    *net.Dialer
	EnableLog bool
}

func (h *Handler) HandleConn(c net.Conn) error {
	defer c.Close()
	br := bufio.NewReader(c)

	// VLESS request basic parse (compat with common clients)
	ver, err := br.ReadByte()
	if err != nil {
		return err
	}
	if ver != 0 {
		return fmt.Errorf("VLESS: unsupported version %d", ver)
	}

	var uuid [16]byte
	if _, err := io.ReadFull(br, uuid[:]); err != nil {
		return err
	}
	if !h.Allow[uuid] {
		return errors.New("VLESS: unauthorized UUID")
	}

	// optLen + opts (we skip)
	optLen, err := br.ReadByte()
	if err != nil {
		return err
	}
	if optLen > 0 {
		if _, err := io.CopyN(io.Discard, br, int64(optLen)); err != nil {
			return err
		}
	}

	// command
	cmd, err := br.ReadByte()
	if err != nil {
		return err
	}
	if cmd != 0x01 { // TCP connect only
		return fmt.Errorf("VLESS: unsupported cmd %d", cmd)
	}

	// reserved 2 bytes
	if _, err := io.CopyN(io.Discard, br, 2); err != nil {
		return err
	}

	// address
	atyp, err := br.ReadByte()
	if err != nil {
		return err
	}
	var host string
	switch atyp {
	case 0x01: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(br, ip); err != nil {
			return err
		}
		host = net.IP(ip).String()
	case 0x02: // domain
		l, err := br.ReadByte()
		if err != nil {
			return err
		}
		b := make([]byte, int(l))
		if _, err := io.ReadFull(br, b); err != nil {
			return err
		}
		host = string(b)
	case 0x04: // IPv6
		ip := make([]byte, 16)
		if _, err := io.ReadFull(br, ip); err != nil {
			return err
		}
		host = net.IP(ip).String()
	default:
		return fmt.Errorf("VLESS: bad atyp %d", atyp)
	}

	// port
	var pbuf [2]byte
	if _, err := io.ReadFull(br, pbuf[:]); err != nil {
		return err
	}
	port := binary.BigEndian.Uint16(pbuf[:])

	if h.EnableLog {
		log.Printf("VLESS req uuid=%x host=%s port=%d", uuid, host, port)
	}

	// Dial out
	target, err := h.Dialer.Dial("tcp", net.JoinHostPort(host, fmt.Sprint(port)))
	if err != nil {
		return fmt.Errorf("dial %s:%d failed: %w", host, port, err)
	}
	defer target.Close()

	// Splice: remaining buffered bytes first
	if br.Buffered() > 0 {
		if _, err := io.CopyN(target, br, int64(br.Buffered())); err != nil {
			return err
		}
	}

	// bidirectional copy
	errc := make(chan error, 2)
	go func() {
		_, e := io.Copy(target, c)
		_ = target.(*net.TCPConn).CloseWrite()
		errc <- e
	}()
	go func() {
		_, e := io.Copy(c, target)
		_ = c.(*net.TCPConn).CloseWrite()
		errc <- e
	}()

	select {
	case <-time.After(24 * time.Hour):
		return nil
	case e := <-errc:
		return e
	}
}
