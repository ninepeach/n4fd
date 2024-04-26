package tlshs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type ExtensionType uint16

// String method for a TLS Extension
// See: http://www.iana.org/assignments/tls-extensiontype-values/tls-extensiontype-values.xhtml
func (t ExtensionType) String() string {
	if name, ok := ExtensionReg[t]; ok {
		return name
	}
	return fmt.Sprintf("%#v (unknown)", t)
}

// TLS Extensions http://www.iana.org/assignments/tls-extensiontype-values/tls-extensiontype-values.xhtml
const (
	ExtServerName           ExtensionType = 0x00
	ExtSupportedGroups      ExtensionType = 0x0a
	ExtECPointFormats       ExtensionType = 0x0b
	ExtSignatureAlgorithms  ExtensionType = 0x0d
	ExtEncryptThenMac       ExtensionType = 0x16
	ExtExtendedMasterSecret ExtensionType = 0x17
	ExtSessionTicket        ExtensionType = 0x23
	ExtKeyShare             ExtensionType = 0x33
	ExtRenegotiationInfo    ExtensionType = 0xff01
)

var ExtensionReg = map[ExtensionType]string{
	ExtServerName:           "server_name",
	ExtSupportedGroups:      "supported_groups",
	ExtECPointFormats:       "ec_point_formats",
	ExtSignatureAlgorithms:  "signature_algorithms_cert",
	ExtEncryptThenMac:       "encrypt_then_mac",
	ExtExtendedMasterSecret: "extended_master_secret",
	ExtSessionTicket:        "SessionTicket TLS",
	ExtKeyShare:             "key_share",
	ExtRenegotiationInfo:    "renegotiation_info",
}

var (
	ErrShortBuffer  = errors.New("short buffer")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrUnknownExt   = errors.New("unknown Extension")
)

type Extension interface {
	Type() ExtensionType
	Encode() ([]byte, error)
	Decode([]byte) error
}

func NewExtension(t ExtensionType, data []byte) (ext Extension, err error) {
	switch t {
	case ExtServerName:
		ext = new(ServerNameExtension)
	case ExtSupportedGroups:
		ext = new(SupportedGroupsExtension)
	case ExtECPointFormats:
		ext = new(ECPointFormatsExtension)
	case ExtSignatureAlgorithms:
		ext = new(SignatureAlgorithmsExtension)
	case ExtEncryptThenMac:
		ext = new(EncryptThenMacExtension)
	case ExtExtendedMasterSecret:
		ext = new(ExtendedMasterSecretExtension)
	case ExtSessionTicket:
		ext = new(SessionTicketExtension)
	case ExtKeyShare:
		ext = new(KeyShareExtension)
	case ExtRenegotiationInfo:
		ext = new(RenegotiationInfoExtension)
	default:
		return nil, ErrUnknownExt
	}
	err = ext.Decode(data)
	return
}

func ParseExtensions(buf []byte, bufLen int) (exts []Extension, err error) {

	var bi int // buf index
	exts = make([]Extension, 0)
	for bi = 0; bi < bufLen; {

		extType := (ExtensionType(buf[bi]) << 8) + ExtensionType(buf[bi+1])
		extLen := int((uint16(buf[bi+2]) << 8) + uint16(buf[bi+3]))
		bi += 4

		ex, _ := NewExtension(extType, buf[bi:bi+extLen])
		if ex != nil {
			exts = append(exts, ex)
		}
		bi += extLen

	}
	return exts, nil
}

type ServerNameExtension struct {
	NameType uint8
	Name     string
}

func (ext *ServerNameExtension) Type() ExtensionType {
	return ExtServerName
}

func (ext *ServerNameExtension) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint16(1+2+len(ext.Name)))
	buf.WriteByte(ext.NameType)
	binary.Write(buf, binary.BigEndian, uint16(len(ext.Name)))
	buf.WriteString(ext.Name)
	return buf.Bytes(), nil
}

func (ext *ServerNameExtension) Decode(b []byte) error {
	if len(b) < 5 {
		return ErrShortBuffer
	}

	ext.NameType = b[2]
	n := int(binary.BigEndian.Uint16(b[3:]))
	if len(b[5:]) < n {
		return ErrShortBuffer
	}
	ext.Name = string(b[5 : 5+n])
	return nil
}

type ECPointFormatsExtension struct {
	Formats []uint8
}

func (ext *ECPointFormatsExtension) Type() ExtensionType {
	return ExtECPointFormats
}

func (ext *ECPointFormatsExtension) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(uint8(len(ext.Formats)))
	buf.Write(ext.Formats)
	return buf.Bytes(), nil
}

func (ext *ECPointFormatsExtension) Decode(b []byte) error {
	if len(b) < 1 {
		return ErrShortBuffer
	}

	n := int(b[0])
	if len(b[1:]) < n {
		return ErrShortBuffer
	}

	ext.Formats = make([]byte, n)
	copy(ext.Formats, b[1:])
	return nil
}

type SupportedGroupsExtension struct {
	Groups []uint16
}

func (ext *SupportedGroupsExtension) Type() ExtensionType {
	return ExtSupportedGroups
}

func (ext *SupportedGroupsExtension) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint16(len(ext.Groups)*2))
	for _, group := range ext.Groups {
		binary.Write(buf, binary.BigEndian, group)
	}
	return buf.Bytes(), nil
}

func (ext *SupportedGroupsExtension) Decode(b []byte) error {
	if len(b) < 2 {
		return ErrShortBuffer
	}

	n := int(binary.BigEndian.Uint16(b)) / 2 * 2 //make it even
	if len(b[2:]) < n {
		return ErrShortBuffer
	}

	for i := 0; i < n; i += 2 {
		ext.Groups = append(ext.Groups, binary.BigEndian.Uint16(b[2+i:]))
	}
	return nil
}

type SignatureAlgorithmsExtension struct {
	Algorithms []uint16
}

func (ext *SignatureAlgorithmsExtension) Type() ExtensionType {
	return ExtSignatureAlgorithms
}

func (ext *SignatureAlgorithmsExtension) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint16(len(ext.Algorithms)*2))
	for _, alg := range ext.Algorithms {
		binary.Write(buf, binary.BigEndian, alg)
	}
	return buf.Bytes(), nil
}

func (ext *SignatureAlgorithmsExtension) Decode(b []byte) error {
	if len(b) < 2 {
		return ErrShortBuffer
	}

	n := int(binary.BigEndian.Uint16(b))
	if len(b[2:]) < n {
		return ErrShortBuffer
	}

	for i := 0; i < n; i += 2 {
		ext.Algorithms = append(ext.Algorithms, binary.BigEndian.Uint16(b[2+i:]))
	}
	return nil
}

type EncryptThenMacExtension struct {
	Data []byte
}

func (ext *EncryptThenMacExtension) Type() ExtensionType {
	return ExtEncryptThenMac
}

func (ext *EncryptThenMacExtension) Encode() ([]byte, error) {
	return ext.Data, nil
}

func (ext *EncryptThenMacExtension) Decode(b []byte) error {
	ext.Data = make([]byte, len(b))
	copy(ext.Data, b)
	return nil
}

type ExtendedMasterSecretExtension struct {
	Data []byte
}

func (ext *ExtendedMasterSecretExtension) Type() ExtensionType {
	return ExtExtendedMasterSecret
}

func (ext *ExtendedMasterSecretExtension) Encode() ([]byte, error) {
	return ext.Data, nil
}

func (ext *ExtendedMasterSecretExtension) Decode(b []byte) error {
	ext.Data = make([]byte, len(b))
	copy(ext.Data, b)
	return nil
}

type SessionTicketExtension struct {
	Data []byte
}

func (ext *SessionTicketExtension) Type() ExtensionType {
	return ExtSessionTicket
}

func (ext *SessionTicketExtension) Encode() ([]byte, error) {
	return ext.Data, nil
}

func (ext *SessionTicketExtension) Decode(b []byte) error {
	ext.Data = make([]byte, len(b))
	copy(ext.Data, b)
	return nil
}

type KeyShareExtension struct {
	Data []byte
}

func (ext *KeyShareExtension) Type() ExtensionType {
	return ExtKeyShare
}

func (ext *KeyShareExtension) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(uint8(len(ext.Data)))
	buf.Write(ext.Data)
	return buf.Bytes(), nil
}

func (ext *KeyShareExtension) Decode(b []byte) error {
	if len(b) < 1 {
		return ErrShortBuffer
	}

	n := int(b[0])
	if len(b[1:]) < n {
		return ErrShortBuffer
	}
	ext.Data = make([]byte, n)
	copy(ext.Data, b[1:])

	return nil
}

type RenegotiationInfoExtension struct {
	Data []byte
}

func (ext *RenegotiationInfoExtension) Type() ExtensionType {
	return ExtRenegotiationInfo
}

func (ext *RenegotiationInfoExtension) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteByte(uint8(len(ext.Data)))
	buf.Write(ext.Data)
	return buf.Bytes(), nil
}

func (ext *RenegotiationInfoExtension) Decode(b []byte) error {
	if len(b) < 1 {
		return ErrShortBuffer
	}

	n := int(b[0])
	if len(b[1:]) < n {
		return ErrShortBuffer
	}
	ext.Data = make([]byte, n)
	copy(ext.Data, b[1:])

	return nil
}
