package tlshs

import (
	"encoding/binary"
	"errors"
)

const (
	HandshakeHeaderSize uint32 = 4
)

const (
	ClientHelloMsgType uint8 = 0x1
	ServerHelloMsgType uint8 = 0x2
)

type Random struct {
	Time   uint32
	Opaque [28]byte
}

type ClientHelloMsg struct {
	MsgType            uint8
	MsgLen             uint16
	TLSVersion         Version
	Random             Random
	SessionIDLen       uint8
	SessionID          []byte
	CipherSuiteLen     uint16
	CipherSuites       []uint16
	CompressionMethods []uint8
	Extensions         []Extension
}

func ParseClientHelloMsg(buf []byte) (hm *ClientHelloMsg, err error) {
	//check buf size > HandshakeHeaderSize
	if len(buf) < int(HandshakeHeaderSize) {
		// must be able to, at least, read the HandshakeHeader
		return nil, errors.New("unsupported handshake message size")
	}

	hm = &ClientHelloMsg{}

	// Handshake Header:
	hm.MsgType = uint8(buf[0])
	if hm.MsgType != ClientHelloMsgType {
		return nil, errors.New("not a client hello handshake message")
	}
	hm.MsgLen = uint16(buf[1])<<16 + uint16(buf[2])<<8 + uint16(buf[3])

	//buf index
	bi := int(HandshakeHeaderSize)
	//check buf size > msg size
	if hm.MsgLen > uint16(len(buf[bi:])) {
		return nil, errors.New("client hello message has invalid length")
	}

	// TLSVersion:
	if len(buf[bi:]) < 2 {
		return nil, errors.New("client hello message has invalid format tls")
	}
	hm.TLSVersion = (Version(buf[bi]) << 8) + Version(buf[bi+1])
	bi += 2

	// Random Time
	if len(buf[bi:]) < 32 {
		return nil, errors.New("client hello message has invalid format random")
	}
	hm.Random.Time = binary.BigEndian.Uint32(buf[bi : bi+4])
	bi += 4
	// Random Data
	bi += copy(hm.Random.Opaque[:], buf[bi:bi+28])

	// SessionID:
	if len(buf[bi:]) < 1 {
		return nil, errors.New("client hello message has invalid format sessionID")
	}
	hm.SessionIDLen = uint8(buf[bi])
	bi += 1
	if len(buf[bi:]) < 1 {
		return nil, errors.New("client hello message has invalid sessionID length")
	}
	hm.SessionID = make([]byte, hm.SessionIDLen)
	bi += copy(hm.SessionID[:], buf[bi:bi+int(hm.SessionIDLen)])

	// CipherSuites:
	if len(buf[bi:]) < 2 {
		return nil, errors.New("client hello message has invalid format cipher suite")
	}
	hm.CipherSuiteLen = (uint16(buf[bi]) << 8) + uint16(buf[bi+1])
	bi += 2

	if len(buf[bi:]) < int(hm.CipherSuiteLen) {
		return nil, errors.New("client hello message has invalid cipher suite length")
	}
	hm.CipherSuites, err = ParseCipherSuites(buf[bi : bi+(int(hm.CipherSuiteLen))])
	if err != nil {
		return nil, err
	}
	bi += int(hm.CipherSuiteLen)

	// CompressionMethods:
	if len(buf[bi:]) < 1 {
		return nil, errors.New("client hello message has invalid format CompressionMethods")
	}
	cmLen := int(buf[bi])
	bi += 1

	for i := 0; i < cmLen; i++ {
		hm.CompressionMethods = append(hm.CompressionMethods, buf[bi+i])
	}
	bi += cmLen

	// Extensions:
	if len(buf[bi:]) < 2 {
		return nil, errors.New("client hello message has invalid format Extensions")
	}
	extsLen := (uint16(buf[bi]) << 8) + uint16(buf[bi+1])
	bi += 2

	if len(buf[bi:]) < int(extsLen) {
		return nil, errors.New("client hello message has invalid format ExtensionsData")
	}
	ExtensionData := make([]byte, extsLen)
	bi += copy(ExtensionData[:], buf[bi:bi+int(extsLen)])

	hm.Extensions, _ = ParseExtensions(ExtensionData, int(extsLen))

	return hm, nil
}

func (hm *ClientHelloMsg) ToBinary() []byte {
	// pre-allocate if length is known, else cap is HandshakeHeaderSize
	raw := make([]byte, 0, uint(hm.MsgLen)+uint(HandshakeHeaderSize))

	raw = append(raw, byte(hm.MsgType))
	raw = append(raw, byte(hm.MsgLen>>16), byte(hm.MsgLen>>8), byte(hm.MsgLen))
	raw = append(raw, byte(hm.TLSVersion))

	randomTime := [4]byte{}
	randomTime[0] = byte(hm.Random.Time >> 24)
	randomTime[1] = byte(hm.Random.Time << 8 >> 24)
	randomTime[2] = byte(hm.Random.Time << 16 >> 24)
	randomTime[3] = byte(hm.Random.Time << 24 >> 24)

	raw = append(raw, randomTime[:]...)
	raw = append(raw, hm.Random.Opaque[:28]...)

	raw = append(raw, hm.SessionIDLen)
	raw = append(raw, hm.SessionID[:]...)
	raw = append(raw, byte(hm.CipherSuiteLen>>8), byte(hm.CipherSuiteLen))
	for i := 0; i < len(hm.CipherSuites); i++ {
		cs := hm.CipherSuites[i]
		raw = append(raw, byte(cs>>8), byte(cs))
	}
	raw = append(raw, hm.CompressionMethods[:]...)

	//extensions to binary
	extData := make([]byte, 0)
	for _, ext := range hm.Extensions {
		var b []byte
		b, err := ext.Encode()
		if err != nil {
			break
		}
		extType := ext.Type()
		extLen := len(b)
		extData = append(extData, byte(extType>>8), byte(extType))
		extData = append(extData, byte(extLen>>8), byte(extLen))
		extData = append(extData, b[:]...)
	}
	raw = append(raw, byte(len(extData)>>8), byte(len(extData)))
	raw = append(raw, extData[:]...)

	hm.MsgLen = uint16(len(raw)) - uint16(HandshakeHeaderSize)
	raw[1] = byte(hm.MsgLen >> 16)
	raw[2] = byte(hm.MsgLen >> 8)
	raw[3] = byte(hm.MsgLen)

	return raw
}

type ServerHelloMsg struct {
	MsgType           uint8
	MsgLen            uint16
	TLSVersion        Version
	Random            Random
	SessionIDLen      uint8
	SessionID         []byte
	CipherSuite       uint16
	CompressionMethod uint8
	Extensions        []Extension
}

func ParseServerHelloMsg(buf []byte) (hm *ServerHelloMsg, err error) {
	if len(buf) < int(HandshakeHeaderSize) {
		// must be able to, at least, read the HandshakeHeader
		return nil, errors.New("unsupported handshake message size")
	}

	hm = &ServerHelloMsg{}

	// Handshake Header:
	hm.MsgType = buf[0]
	if hm.MsgType != ServerHelloMsgType {
		return nil, errors.New("not a server hello handshake message")
	}
	hm.MsgLen = uint16(buf[1])<<16 + uint16(buf[2])<<8 + uint16(buf[3])

	//buf index
	bi := int(HandshakeHeaderSize)
	if hm.MsgLen > uint16(len(buf[bi:])) {
		return nil, errors.New("server hello message has invalid length")
	}

	// TLSVersion:
	if len(buf[bi:]) < 2 {
		return nil, errors.New("server hello message has invalid format tls")
	}
	hm.TLSVersion = (Version(buf[bi]) << 8) + Version(buf[bi+1])
	bi += 2

	// Random Time
	if len(buf[bi:]) < 32 {
		return nil, errors.New("server hello message has invalid format random")
	}
	hm.Random.Time = binary.BigEndian.Uint32(buf[bi : bi+4])
	bi += 4
	// Random Data
	bi += copy(hm.Random.Opaque[:], buf[bi:bi+28])

	// SessionID:
	if len(buf[bi:]) < 1 {
		return nil, errors.New("server hello message has invalid format sessionID")
	}
	hm.SessionIDLen = uint8(buf[bi])
	bi += 1

	if len(buf[bi:]) < int(hm.SessionIDLen) || int(hm.SessionIDLen) < 1 {
		return nil, errors.New("server hello message has invalid sessionID length")
	}
	hm.SessionID = make([]byte, hm.SessionIDLen)
	bi += copy(hm.SessionID[:], buf[bi:bi+int(hm.SessionIDLen)])

	// CipherSuite:
	if len(buf[bi:]) < 2 {
		return nil, errors.New("server hello message has invalid format Cipher Suite")
	}
	hm.CipherSuite = (uint16(buf[bi]) << 8) + uint16(buf[bi+1])
	bi += 2

	// CompressionMethod:
	if len(buf[bi:]) < 1 {
		return nil, errors.New("server hello message has invalid format CompressionMethod")
	}
	hm.CompressionMethod = uint8(buf[bi])
	bi += 1

	// Extensions:
	if len(buf[bi:]) < 2 {
		return nil, errors.New("server hello message has invalid format Extensions")
	}
	extsLen := (uint16(buf[bi]) << 8) + uint16(buf[bi+1])
	bi += 2

	if len(buf[bi:]) < int(extsLen) {
		return nil, errors.New("server hello message has invalid format ExtensionsData")
	}
	ExtensionData := make([]byte, extsLen)
	bi += copy(ExtensionData[:], buf[bi:bi+int(extsLen)])

	hm.Extensions, _ = ParseExtensions(ExtensionData, int(extsLen))

	return hm, nil
}

func (hm *ServerHelloMsg) ToBinary() []byte {
	// Pre-allocate if length is known, else cap is HandshakeHeaderByteSize
	raw := make([]byte, 0, hm.MsgLen+uint16(HandshakeHeaderSize))

	hm.MsgType = ServerHelloMsgType
	raw = append(raw, byte(hm.MsgType))
	raw = append(raw, byte(hm.MsgLen>>16), byte(hm.MsgLen>>8), byte(hm.MsgLen))
	raw = append(raw, byte(hm.TLSVersion>>8), byte(hm.TLSVersion))

	randomTime := [4]byte{}
	randomTime[0] = byte(hm.Random.Time >> 24)
	randomTime[1] = byte(hm.Random.Time << 8 >> 24)
	randomTime[2] = byte(hm.Random.Time << 16 >> 24)
	randomTime[3] = byte(hm.Random.Time << 24 >> 24)

	raw = append(raw, randomTime[:]...)
	raw = append(raw, hm.Random.Opaque[:28]...)

	hm.SessionIDLen = uint8(len(hm.SessionID))
	raw = append(raw, byte(hm.SessionIDLen))
	raw = append(raw, hm.SessionID[:]...)

	raw = append(raw, byte(hm.CipherSuite>>8), byte(hm.CipherSuite))
	raw = append(raw, byte(hm.CompressionMethod))

	//extensions to binary
	extData := make([]byte, 0)
	for _, ext := range hm.Extensions {
		var b []byte
		b, err := ext.Encode()
		if err != nil {
			break
		}
		extType := ext.Type()
		extLen := len(b)
		extData = append(extData, byte(extType>>8), byte(extType))
		extData = append(extData, byte(extLen>>8), byte(extLen))
		extData = append(extData, b[:]...)
	}
	raw = append(raw, byte(len(extData)>>8), byte(len(extData)))
	raw = append(raw, extData[:]...)

	hm.MsgLen = uint16(len(raw)) - uint16(HandshakeHeaderSize)
	raw[1] = byte(hm.MsgLen >> 16)
	raw[2] = byte(hm.MsgLen >> 8)
	raw[3] = byte(hm.MsgLen)

	return raw
}
