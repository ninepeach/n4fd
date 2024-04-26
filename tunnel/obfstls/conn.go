package obfstls

import (
	"bytes"
	"github.com/ninepeach/n4fd/log"
	"github.com/ninepeach/n4fd/tunnel"
	"github.com/ninepeach/n4fd/tunnel/obfstls/tlshs"
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

	//err = tlshs.HandshakeWithClient(c.Conn)
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
	if record.RecordType != tlshs.HandshakeRecord {
		return tlshs.ErrHandShake
	}

	clientMsg, _ := tlshs.ParseClientHelloMsg(record.Data)

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
		TLSVersion:        tlshs.VersionTLS12,
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
	rand.Read(serverMsg.Random.Data[:])
	b := serverMsg.ToBinary()

	record = &tlshs.Record{
		RecordType: tlshs.HandshakeRecord,
		TLSVersion: tlshs.VersionTLS10,
		Length:     uint16(len(b)),
		Data:       b,
	}

	if _, err := record.WriteTo(&c.wbuf); err != nil {
		return err
	}

	record = &tlshs.Record{
		RecordType: tlshs.ChangeCipherRecord,
		TLSVersion: tlshs.VersionTLS12,
		Length:     uint16(1),
		Data:       []byte{0x01},
	}
	if _, err := record.WriteTo(&c.wbuf); err != nil {
		return err
	}
	return nil
}

func (c *Conn) Read(b []byte) (n int, err error) {

	if c.rbuf.Len() > 0 {
		return c.rbuf.Read(b)
	}
	record := &tlshs.Record{}
	if _, err = record.ReadFrom(c.Conn); err != nil {
		return
	}
	n = copy(b, record.Data)
	_, err = c.rbuf.Write(record.Data[n:])

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
			RecordType: tlshs.ApplicationRecord,
			TLSVersion: tlshs.VersionTLS12,
			Length:     uint16(len(data)),
			Data:       data,
		}

		if c.wbuf.Len() > 0 {
			record.RecordType = tlshs.HandshakeRecord
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
