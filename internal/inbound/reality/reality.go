package reality

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/xtls/reality"
)

type NextHandler interface {
	HandleConn(net.Conn) error
}

type Inbound struct {
	// runtime
	Listen string

	// REALITY verification
	PrivKey  []byte            // 32 bytes
	SNIAllow map[string]bool   // nil => allow any SNI
	ShortIDs map[[8]byte]bool  // allow empty if contains zero key
	Dest     string            // e.g., "13.107.21.200:443" (recommended IP:443)
	Debug    bool

	// next stage (VLESS)
	Next NextHandler
}

func (in *Inbound) Serve(ln net.Listener) error {
	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil {
			return err
		}
		go in.handleConn(c)
	}
}

func (in *Inbound) handleConn(c net.Conn) {
	defer func() {
		_ = c.Close()
	}()

	remote := c.RemoteAddr().String()
	if in.Debug {
		log.Printf("DEBUG TCP prelude remote=%s", remote)
	}

	rc := &reality.Config{
		// Important dialing: REALITY will contact the "dest" site to obtain legit server hello etc.
		DialContext: (&net.Dialer{Timeout: 10 * time.Second}).DialContext,

		Show: true,
		Type: "tcp",
		Dest: in.Dest,
		Xver: 0,

		PrivateKey:  in.PrivKey,
		MaxTimeDiff: 2 * time.Minute,

		// MUST keep these as shown by REALITY docs:
		NextProtos:             nil,
		SessionTicketsDisabled: true,
	}

	// SNI allow list
	if in.SNIAllow != nil {
		rc.ServerNames = in.SNIAllow
	} else {
		rc.ServerNames = map[string]bool{} // empty map means no filter inside; we do not block at this layer
	}

	// Short IDs
	rc.ShortIds = make(map[[8]byte]bool, len(in.ShortIDs))
	for k := range in.ShortIDs {
		rc.ShortIds[k] = true
	}

	// run REALITY handshake
	conn, err := safeRealityServer(context.Background(), c, rc)
	if err != nil {
		if in.Debug {
			log.Printf("DEBUG reality.Server failed/aborted remote=%s err=%v", remote, err)
		}
		return
	}

	// at this point, REALITY verified; hand off to next
	if in.Next == nil {
		if in.Debug {
			log.Printf("WARN no Next handler set; closing")
		}
		return
	}
	if in.Debug {
		log.Printf("DEBUG REALITY auth passed; dispatch to VLESS for %s", remote)
	}
	_ = in.Next.HandleConn(conn)
}

func safeRealityServer(ctx context.Context, c net.Conn, rc *reality.Config) (_ *reality.Conn, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("reality handshake aborted")
		}
	}()
	return reality.Server(ctx, c, rc)
}
