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
	Time uint32
	Data [28]byte
}

type ClientHelloMsg struct {
	Type               uint8
	Length             uint16
	TLSVersion         uint16
	Random             Random
	SessionIDLen       uint8
	SessionID          []byte
	CipherSuiteLen     uint16
	CipherSuite        []uint16
	CompressionMethods []byte
	ExtensionsLen      uint16
	ExtensionData      []byte
	Extensions         []Extension
}

func ParseClientHelloMsg(buf []byte) (hm *ClientHelloMsg, err error) {
	if len(buf) < int(HandshakeHeaderSize) {
		// must be able to, at least, read the HandshakeHeader
		return nil, errors.New("unsupported handshake message size")
	}

	hm = &ClientHelloMsg{}

	// Handshake Header:
	hm.Type = uint8(buf[0])
	if hm.Type != ClientHelloMsgType {
		return nil, errors.New("not a client hello handshake message")
	}
	hm.Length = uint16(buf[1])<<16 + uint16(buf[2])<<8 + uint16(buf[3])

	//buf index
	bi := int(HandshakeHeaderSize)
	if hm.Length > uint16(len(buf[bi:])) {
		return nil, errors.New("client hello message has invalid length")
	}

	// TLSVersion:
	if len(buf[bi:]) < 2 {
		return nil, errors.New("client hello message has invalid format tls")
	}
	hm.TLSVersion = binary.BigEndian.Uint16(buf[bi : bi+2])
	bi += 2

	// Random Time
	if len(buf[bi:]) < 32 {
		return nil, errors.New("server hello message has invalid format random")
	}
	hm.Random.Time = binary.BigEndian.Uint32(buf[bi : bi+4])
	bi += 4
	// Random Data
	bi += copy(hm.Random.Data[:], buf[bi:bi+28])

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

	// CipherSuite:
	if len(buf[bi:]) < 2 {
		return nil, errors.New("client hello message has invalid format cipher suite")
	}
	hm.CipherSuiteLen = (uint16(buf[bi]) << 8) + uint16(buf[bi+1])
	bi += 2

	if len(buf[bi:]) < int(hm.CipherSuiteLen) {
		return nil, errors.New("client hello message has invalid cipher suite length")
	}
	hm.CipherSuite, err = ParseCipherSuites(buf[bi : bi+(int(hm.CipherSuiteLen))])
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
	hm.ExtensionsLen = (uint16(buf[bi]) << 8) + uint16(buf[bi+1])
	bi += 2

	if len(buf[bi:]) < int(hm.ExtensionsLen) {
		return nil, errors.New("client hello message has invalid format ExtensionsData")
	}
	hm.ExtensionData = make([]byte, hm.ExtensionsLen)
	bi += copy(hm.ExtensionData[:], buf[bi:bi+int(hm.ExtensionsLen)])

	hm.Extensions, _ = ParseExtensions(hm.ExtensionData, int(hm.ExtensionsLen))

	return hm, nil
}

func (hm *ClientHelloMsg) ToBinary() []byte {
	// pre-allocate if length is known, else cap is HandshakeHeaderSize
	raw := make([]byte, 0, uint(hm.Length)+uint(HandshakeHeaderSize))

	raw = append(raw, byte(hm.Type))
	raw = append(raw, byte(hm.Length>>16), byte(hm.Length>>8), byte(hm.Length))
	raw = append(raw, byte(hm.TLSVersion))

	randomTime := [4]byte{}
	randomTime[0] = byte(hm.Random.Time >> 24)
	randomTime[1] = byte(hm.Random.Time << 8 >> 24)
	randomTime[2] = byte(hm.Random.Time << 16 >> 24)
	randomTime[3] = byte(hm.Random.Time << 24 >> 24)

	raw = append(raw, randomTime[:]...)
	raw = append(raw, hm.Random.Data[:28]...)

	raw = append(raw, hm.SessionIDLen)
	raw = append(raw, hm.SessionID[:]...)
	raw = append(raw, byte(hm.CipherSuiteLen>>8), byte(hm.CipherSuiteLen))
	for i := 0; i < len(hm.CipherSuite); i++ {
		cs := hm.CipherSuite[i]
		raw = append(raw, byte(cs>>8), byte(cs))
	}
	raw = append(raw, hm.CompressionMethods[:]...)
	raw = append(raw, byte(hm.ExtensionsLen>>8), byte(hm.ExtensionsLen))
	raw = append(raw, hm.ExtensionData[:]...)

	if hm.Length == 0 {
		// Automatically figure out the length.
		// Length was previously written as 0, now it's calculated and we need to update it.
		hm.Length = uint16(len(raw)) - uint16(HandshakeHeaderSize)
		raw[1] = byte(hm.Length >> 16)
		raw[2] = byte(hm.Length >> 8)
		raw[3] = byte(hm.Length)
	}

	return raw
}

type ServerHelloMsg struct {
	Type              uint8
	Length            uint16
	TLSVersion        uint16
	Random            Random
	SessionIDLen      uint8
	SessionID         []byte
	CipherSuite       uint16
	CompressionMethod uint8
	ExtensionsLen     uint16
	ExtensionData     []byte
	Extensions        []Extension
}

func ParseServerHelloMsg(buf []byte) (hm *ServerHelloMsg, err error) {
	if len(buf) < int(HandshakeHeaderSize) {
		// must be able to, at least, read the HandshakeHeader
		return nil, errors.New("unsupported handshake message size")
	}

	hm = &ServerHelloMsg{}

	// Handshake Header:
	hm.Type = buf[0]
	if hm.Type != ServerHelloMsgType {
		return nil, errors.New("not a server hello handshake message")
	}
	hm.Length = uint16(buf[1])<<16 + uint16(buf[2])<<8 + uint16(buf[3])

	if hm.Length > uint16(len(buf[4:])) {
		return nil, errors.New("server hello message has invalid length")
	}

	//buf index
	bi := int(HandshakeHeaderSize)
	if hm.Length > uint16(len(buf[bi:])) {
		return nil, errors.New("server hello message has invalid length")
	}

	// TLSVersion:
	if len(buf[bi:]) < 2 {
		return nil, errors.New("server hello message has invalid format tls")
	}
	hm.TLSVersion = binary.BigEndian.Uint16(buf[bi : bi+2])
	bi += 2

	// Random Time
	if len(buf[bi:]) < 32 {
		return nil, errors.New("server hello message has invalid format random")
	}
	hm.Random.Time = binary.BigEndian.Uint32(buf[bi : bi+4])
	bi += 4
	// Random Data
	bi += copy(hm.Random.Data[:], buf[bi:bi+28])

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
	hm.ExtensionsLen = (uint16(buf[bi]) << 8) + uint16(buf[bi+1])
	bi += 2

	if len(buf[bi:]) < int(hm.ExtensionsLen) {
		return nil, errors.New("server hello message has invalid format ExtensionsData")
	}
	hm.ExtensionData = make([]byte, hm.ExtensionsLen)
	bi += copy(hm.ExtensionData[:], buf[bi:bi+int(hm.ExtensionsLen)])

	hm.Extensions, _ = ParseExtensions(hm.ExtensionData, int(hm.ExtensionsLen))

	return hm, nil
}

func (hm *ServerHelloMsg) ToBinary() []byte {
	// Pre-allocate if length is known, else cap is HandshakeHeaderByteSize
	raw := make([]byte, 0, hm.Length+uint16(HandshakeHeaderSize))

	hm.Type = ServerHelloMsgType
	raw = append(raw, byte(hm.Type))
	raw = append(raw, byte(hm.Length>>16), byte(hm.Length>>8), byte(hm.Length))
	raw = append(raw, byte(hm.TLSVersion>>8), byte(hm.TLSVersion))

	randomTime := [4]byte{}
	randomTime[0] = byte(hm.Random.Time >> 24)
	randomTime[1] = byte(hm.Random.Time << 8 >> 24)
	randomTime[2] = byte(hm.Random.Time << 16 >> 24)
	randomTime[3] = byte(hm.Random.Time << 24 >> 24)

	raw = append(raw, randomTime[:]...)
	raw = append(raw, hm.Random.Data[:28]...)

	hm.SessionIDLen = uint8(len(hm.SessionID))
	raw = append(raw, byte(hm.SessionIDLen))
	raw = append(raw, hm.SessionID[:]...)

	raw = append(raw, byte(hm.CipherSuite>>8), byte(hm.CipherSuite))
	raw = append(raw, byte(hm.CompressionMethod))

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

	hm.ExtensionsLen = uint16(len(extData))
	hm.ExtensionData = extData
	raw = append(raw, byte(hm.ExtensionsLen>>8), byte(hm.ExtensionsLen))
	raw = append(raw, hm.ExtensionData[:]...)

	if hm.Length == 0 {
		// Automatically figure out the length.
		// Length was previously written as 0, now it's calculated and we need to update it.
		hm.Length = uint16(len(raw)) - uint16(HandshakeHeaderSize)
		raw[1] = byte(hm.Length >> 16)
		raw[2] = byte(hm.Length >> 8)
		raw[3] = byte(hm.Length)
	} else {
		if hm.Length != uint16(len(raw))-uint16(HandshakeHeaderSize) {
			hm.Length = uint16(len(raw)) - uint16(HandshakeHeaderSize)
			raw = hm.ToBinary()
		}
	}

	return raw
}
