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

	VersionTLS10 uint16 = 0x0301
	VersionTLS11 uint16 = 0x0302
	VersionTLS12 uint16 = 0x0303
	VersionTLS13 uint16 = 0x0304
)

var (
	ErrHandShake = errors.New("handshake failed")
)

type Record struct {
	RecordType uint8
	TLSVersion uint16
	Length     uint16 // bytes in rest of the handshake message
	Data       []byte
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
		record.RecordType = HandshakeRecord
	case AlertRecord:
		record.RecordType = AlertRecord
	case ApplicationRecord:
		record.RecordType = ApplicationRecord
	default:
		return 0, errors.New("unsupported record type")
	}
	record.RecordType = buf[0]
	record.TLSVersion = (uint16(buf[1]) << 8) + uint16(buf[2])

	//support TLS10 TLS11 TLS12 TLS13
	if record.TLSVersion == uint16(VersionTLS13) {
		record.TLSVersion = uint16(VersionTLS13)
	} else if record.TLSVersion == uint16(VersionTLS12) {
		record.TLSVersion = uint16(VersionTLS12)
	} else if record.TLSVersion == uint16(VersionTLS11) {
		record.TLSVersion = uint16(VersionTLS11)
	} else if record.TLSVersion == uint16(VersionTLS10) {
		record.TLSVersion = uint16(VersionTLS10)
	} else {
		return 0, errors.New("unsupported version of TLS")
	}

	record.Length = (uint16(buf[3]) << 8) + uint16(buf[4])
	if record.Length > MaxRecordSize {
		return 0, errors.New("record length exceeds the maximum for a record")
	}

	record.Data = make([]byte, record.Length)
	nn, err = io.ReadFull(r, record.Data)

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
	raw := make([]byte, RecordHeaderSize)
	raw[0] = byte(record.RecordType)
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
	copy(raw[RecordHeaderSize:], record.Data[:])
	return raw
}
