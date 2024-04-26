package tlshs

import (
	"errors"
	"io"
)

const (
	//Record Header Size
	RecordHeaderSize uint16 = 5
	MaxRecordSize    uint16 = 16384

	//Record Type
	ChangeCipherRecord uint8 = 0x14
	AlertRecord        uint8 = 0x15
	HandshakeRecord    uint8 = 0x16
	ApplicationRecord  uint8 = 0x17
	SSL2Record         uint8 = 0x80
)

type Record struct {
	Type       uint8
	TLSVersion Version
	Length     uint16 // bytes in rest of the handshake message
	Opaque     []byte
}

func (record *Record) ReadFrom(r io.Reader) (n int64, err error) {
	//read header to buf
	buf := make([]byte, RecordHeaderSize)
	nn, err := io.ReadFull(r, buf)
	n += int64(nn)
	if err != nil {
		return
	}

	//check record type
	switch uint8(buf[0]) {
	case HandshakeRecord:
		record.Type = HandshakeRecord
	case AlertRecord:
		record.Type = AlertRecord
	case ApplicationRecord:
		record.Type = ApplicationRecord
	case ChangeCipherRecord:
		record.Type = ChangeCipherRecord
	default:
		return 0, errors.New("unsupported record type")
	}

	//set record type
	record.Type = buf[0]
	record.TLSVersion = (Version(buf[1]) << 8) + Version(buf[2])

	//support TLS10 TLS11 TLS12 TLS13
	if record.TLSVersion == Version(VersionTLS13) {
		record.TLSVersion = Version(VersionTLS13)
	} else if record.TLSVersion == Version(VersionTLS12) {
		record.TLSVersion = Version(VersionTLS12)
	} else if record.TLSVersion == Version(VersionTLS11) {
		record.TLSVersion = Version(VersionTLS11)
	} else if record.TLSVersion == Version(VersionTLS10) {
		record.TLSVersion = Version(VersionTLS10)
	} else {
		return 0, errors.New("unsupported version of TLS")
	}

	record.Length = (uint16(buf[3]) << 8) + uint16(buf[4])
	if record.Length > MaxRecordSize {
		return 0, errors.New("record length exceeds the maximum for a record")
	}

	record.Opaque = make([]byte, record.Length)
	nn, err = io.ReadFull(r, record.Opaque)

	n += int64(nn)
	return
}

func (record *Record) WriteTo(w io.Writer) (n int64, err error) {
	nn, err := w.Write(record.ToBinary())
	n += int64(nn)
	return
}

func (record *Record) Size() int {
	return int(record.Length + RecordHeaderSize)
}

func (record *Record) HeaderToBinary() []byte {
	record.Length = uint16(len(record.Opaque))

	raw := make([]byte, RecordHeaderSize)
	raw[0] = byte(record.Type)
	raw[1] = byte(record.TLSVersion >> 8)
	raw[2] = byte(record.TLSVersion)
	raw[3] = byte(record.Length >> 8)
	raw[4] = byte(record.Length)
	return raw
}

func (record *Record) ToBinary() []byte {
	headerBytes := record.HeaderToBinary()
	raw := make([]byte, record.Size())
	copy(raw[:], headerBytes[:])
	copy(raw[RecordHeaderSize:], record.Opaque[:])
	return raw
}
