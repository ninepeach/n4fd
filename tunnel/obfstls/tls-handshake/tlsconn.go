package tlshs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type writer struct {
	io.Writer
	buf []byte
}

// NewWriter wraps an io.Writer with tls record decrypted.
func NewWriter(w io.Writer) io.Writer { return newWriter(w) }

func newWriter(w io.Writer) *writer {
	return &writer{
		Writer: w,
		buf:    make([]byte, maxTLSDataLen+recordHeaderLen),
	}
}

// Write encrypts b and writes to the embedded io.Writer.
func (w *writer) Write(b []byte) (int, error) {
	n, err := w.ReadFrom(bytes.NewBuffer(b))

	return int(n), err
}

// ReadFrom reads from the given io.Reader until EOF or error, encrypts and
// writes to the embedded io.Writer. Returns number of bytes read from r and
// any error encountered.
func (w *writer) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		buf := w.buf
		payloadBuf := buf[recordHeaderLen : recordHeaderLen+maxTLSDataLen]
		nr, er := r.Read(payloadBuf)

		if nr > 0 {
			n += int64(nr)
			buf = buf[:recordHeaderLen+nr]
			payloadBuf = payloadBuf[:nr]

			record := &Record{
				Type:    RecordTypeAppData,
				Version: VersionTLS10,
				Opaque:  buf,
			}
			_, ew := record.WriteTo(w)
			if ew != nil {
				err = ew
				break
			}
		}

		if er != nil {
			if er != io.EOF { // ignore EOF as per io.ReaderFrom contract
				err = er
			}
			break
		}
	}

	return n, err
}

type reader struct {
	io.Reader
	buf      []byte
	leftover []byte
}

func NewReader(r io.Reader) io.Reader { return newReader(r) }

func newReader(r io.Reader) *reader {
	return &reader{
		Reader: r,
		buf:    make([]byte, maxTLSDataLen),
	}
}

func (r *reader) read() (int, error) {
	recordHeader := make([]byte, 350)
	_, err := io.ReadFull(r.Reader, recordHeader)
	if err != nil {
		return 0, err
	}

	fmt.Println("debug read ", string(recordHeader), recordHeader)
	if recordHeader[0] != 23 {
		return 0, ErrBadRecordType
	}

	dataLen := int(binary.BigEndian.Uint16(recordHeader[3:5]))

	if dataLen > maxTLSDataLen {
		return 0, ErrMaxDataLen
	}

	buf := r.buf[:dataLen]
	_, err = io.ReadFull(r.Reader, buf)
	if err != nil {
		return 0, err
	}

	fmt.Println("debug read ", recordHeader, buf[0:16], buf[16:dataLen], string(buf))
	return dataLen, nil
}

func (r *reader) Read(b []byte) (int, error) {

	if len(r.leftover) > 0 {
		n := copy(b, r.leftover)
		fmt.Println("debug Read b ", n, b[:n])

		r.leftover = r.leftover[n:]
		return n, nil
	}

	n, err := r.read()
	m := copy(b, r.buf[:n])

	if m < n {
		r.leftover = r.buf[m:n]
	}
	fmt.Println("debug Read ", n, m)
	return m, err
}

func (r *reader) WriteTo(w io.Writer) (n int64, err error) {
	for len(r.leftover) > 0 {
		nw, ew := w.Write(r.leftover)
		r.leftover = r.leftover[nw:]
		n += int64(nw)
		if ew != nil {
			return n, ew
		}
	}

	for {
		nr, er := r.read()
		if nr > 0 {
			nw, ew := w.Write(r.buf[:nr])
			n += int64(nw)

			if ew != nil {
				err = ew
				break
			}
		}

		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	return n, err
}

type streamConn struct {
	net.Conn
	r *reader
	w *writer
}

func (c *streamConn) initReader() error {

	//sessionId, err := ReadClientHello(c.Conn)
	//fmt.Println("readClientHello", err)
	//SendServerHello(c.Conn, sessionId)

	c.r = newReader(c.Conn)

	return nil
}

func (c *streamConn) Read(b []byte) (int, error) {

	if c.r == nil {
		if err := c.initReader(); err != nil {
			return 0, err
		}
	}

	return c.r.Read(b)
}

func (c *streamConn) WriteTo(w io.Writer) (int64, error) {
	if c.r == nil {
		if err := c.initReader(); err != nil {
			return 0, err
		}
	}
	return c.r.WriteTo(w)
}

func (c *streamConn) initWriter() error {

	c.w = newWriter(c.Conn)
	return nil
}

func (c *streamConn) Write(b []byte) (int, error) {
	if c.w == nil {
		if err := c.initWriter(); err != nil {
			return 0, err
		}
	}
	return c.w.Write(b)
}

func (c *streamConn) ReadFrom(r io.Reader) (int64, error) {
	if c.w == nil {
		if err := c.initWriter(); err != nil {
			return 0, err
		}
	}
	return c.w.ReadFrom(r)
}

// NewConn wraps a stream-oriented net.Conn
func NewConn(c net.Conn) net.Conn { return &streamConn{Conn: c} }
