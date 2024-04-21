package obfstls

import (
	"bytes"
	"github.com/ninepeach/n4fd/log"
	"github.com/ninepeach/n4fd/tunnel"
	"github.com/ninepeach/n4fd/tunnel/obfstls/tls-handshake"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Conn struct {
	net.Conn
	metadata       *tunnel.Metadata
	rbuf           bytes.Buffer
	wbuf           bytes.Buffer
	isServer       bool
	handshaked     chan struct{}
	handshakeMutex sync.Mutex
}

func (c *Conn) Handshaked() bool {
	select {
	case <-c.handshaked:
		return true
	default:
		return false
	}
}

func (c *Conn) Handshake() (err error) {
	c.handshakeMutex.Lock()
	defer c.handshakeMutex.Unlock()

	if c.Handshaked() {
		return
	}

	err = c.serverHandshake()
	if err != nil {
		return
	}

	close(c.handshaked)
	return nil
}

func (c *Conn) serverHandshake() error {
	record := &tlshs.Record{}
	if _, err := record.ReadFrom(c.Conn); err != nil {
		log.Debug(err)
		return err
	}
	if record.Type != tlshs.RecordTypeHandshake {
		return tlshs.ErrBadRecordType
	}

	clientMsg := &tlshs.ClientHelloMsg{}
	if err := clientMsg.Decode(record.Opaque); err != nil {
		log.Debug(err)
		return err
	}

	for _, ext := range clientMsg.Extensions {
		if ext.Type() == tlshs.ExtSessionTicket {
			b, err := ext.Encode()
			if err != nil {
				log.Debug(err)
				return err
			}
			c.rbuf.Write(b)
			break
		}
	}

	serverMsg := &tlshs.ServerHelloMsg{
		Version:           tlshs.VersionTLS12,
		SessionID:         clientMsg.SessionID,
		CipherSuite:       0xcca8,
		CompressionMethod: 0x00,
		Extensions: []tlshs.Extension{
			&tlshs.RenegotiationInfoExtension{},
			&tlshs.ExtendedMasterSecretExtension{},
			&tlshs.ECPointFormatsExtension{
				Formats: []uint8{0x00},
			},
		},
	}

	serverMsg.Random.Time = uint32(time.Now().Unix())
	rand.Read(serverMsg.Random.Opaque[:])
	b, err := serverMsg.Encode()
	if err != nil {
		return err
	}

	record = &tlshs.Record{
		Type:    tlshs.RecordTypeHandshake,
		Version: tlshs.VersionTLS10,
		Opaque:  b,
	}

	if _, err := record.WriteTo(&c.wbuf); err != nil {
		return err
	}

	record = &tlshs.Record{
		Type:    tlshs.RecordTypeChangeCipher,
		Version: tlshs.VersionTLS12,
		Opaque:  []byte{0x01},
	}
	if _, err := record.WriteTo(&c.wbuf); err != nil {
		return err
	}
	return nil
}

func (c *Conn) Read(b []byte) (n int, err error) {

	if err = c.Handshake(); err != nil {
		return
	}

	select {
	case <-c.handshaked:
	}

	if c.rbuf.Len() > 0 {
		return c.rbuf.Read(b)
	}
	record := &tlshs.Record{}
	if _, err = record.ReadFrom(c.Conn); err != nil {
		return
	}
	n = copy(b, record.Opaque)
	_, err = c.rbuf.Write(record.Opaque[n:])

	return n, err
}

func (c *Conn) Write(b []byte) (n int, err error) {
	n = len(b)
	for len(b) > 0 {
		data := b
		maxTLSDataLen := 16384
		if len(b) > maxTLSDataLen {
			data = b[:maxTLSDataLen]
			b = b[maxTLSDataLen:]
		} else {
			b = b[:0]
		}
		record := &tlshs.Record{
			Type:    tlshs.RecordTypeAppData,
			Version: tlshs.VersionTLS12,
			Opaque:  data,
		}

		if c.wbuf.Len() > 0 {
			record.Type = tlshs.RecordTypeHandshake
			record.WriteTo(&c.wbuf)
			_, err = c.wbuf.WriteTo(c.Conn)
			return
		}

		if _, err = record.WriteTo(c.Conn); err != nil {
			return
		}
	}
	return
}

func (c *Conn) Close() error {

	return c.Conn.Close()
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return nil
}
