package tlshs

import (
	"errors"
)

const ()

const ()

const (
	recordHeaderLen    = 5
	handshakeHeaderLen = 4
	maxTLSDataLen      = 16384
	maxHandshake       = 65536

	VersionTLS10 = 0x0301
	VersionTLS11 = 0x0302
	VersionTLS12 = 0x0303
	VersionTLS13 = 0x0304

	// Deprecated: SSLv3 is cryptographically broken, and is no longer
	// supported by this package. See golang.org/issue/32716.
	VersionSSL30 = 0x0300
)

var (
	ErrBadRecordType    = errors.New("tls bad record type")
	ErrBadSessionIdType = errors.New("tls bad record SessionID")
	ErrMaxDataLen       = errors.New("tls bad data length")
	ErrBadHandShakeType = errors.New("tls bad Handshake type")
)
