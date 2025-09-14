package vless

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/ninepeach/n4fd/internal/util/uuid"
)

const (
	Version0 = 0x00

	CmdTCP = 0x01
	CmdUDP = 0x02
	CmdMux = 0x03
)

type Request struct {
	Version    byte
	UserID     uuid.UUID
	AddonsRaw  []byte // protobuf: 这里只透传，不做强解析；我们只识别 flow 是否包含 "xtls-rprx-vision"
	Command    byte
	TargetAddr string
	TargetPort uint16
	Flow       string // 从 AddonsRaw 里弱解析（可为空）
}

var (
	errBadVersion = errors.New("invalid VLESS version")
	errBadUser    = errors.New("invalid VLESS user")
	errAddr       = errors.New("invalid addr/port")
)

// 最小地址编解码（与 Xray 一致）
const (
	atypIPv4   = 0x01
	atypDomain = 0x02
	atypIPv6   = 0x04
)

func readN(r io.Reader, n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(r, b)
	return b, err
}

func decodeAddrPort(r *bufio.Reader) (host string, port uint16, err error) {
	// AddressType then payload, then port
	t, err := r.ReadByte()
	if err != nil {
		return "", 0, err
	}
	switch t {
	case atypIPv4:
		ip, e := readN(r, 4)
		if e != nil {
			return "", 0, e
		}
		host = net.IP(ip).String()
	case atypIPv6:
		ip, e := readN(r, 16)
		if e != nil {
			return "", 0, e
		}
		host = net.IP(ip).String()
	case atypDomain:
		l, e := r.ReadByte()
		if e != nil {
			return "", 0, e
		}
		if l == 0 {
			return "", 0, errAddr
		}
		d, e := readN(r, int(l))
		if e != nil {
			return "", 0, e
		}
		host = string(d)
	default:
		return "", 0, fmt.Errorf("unknown addr type: 0x%x", t)
	}
	var p2 [2]byte
	if _, err = io.ReadFull(r, p2[:]); err != nil {
		return "", 0, err
	}
	port = binary.BigEndian.Uint16(p2[:])
	return
}

func ParseRequest(br *bufio.Reader, allow func(uuid.UUID) bool) (*Request, error) {
	v, err := br.ReadByte()
	if err != nil {
		return nil, err
	}
	if v != Version0 {
		return nil, errBadVersion
	}

	u16, err := readN(br, 16)
	if err != nil {
		return nil, err
	}
	var uid uuid.UUID
	copy(uid[:], u16)
	if allow != nil && !allow(uid) {
		return nil, errBadUser
	}

	// Addons: 1 byte length, then payload
	al, err := br.ReadByte()
	if err != nil {
		return nil, err
	}
	var addons []byte
	if al > 0 {
		addons, err = readN(br, int(al))
		if err != nil {
			return nil, err
		}
	}

	// 弱解析 flow（proto 本应解析，这里仅做最小容错：找字符串）
	flow := ""
	if len(addons) > 0 {
		s := string(addons)
		if strings.Contains(s, "xtls") && strings.Contains(s, "vision") {
			flow = "xtls-rprx-vision"
		}
	}

	// Command
	cmd, err := br.ReadByte()
	if err != nil {
		return nil, err
	}

	var host string
	var port uint16
	switch cmd {
	case CmdTCP, CmdUDP:
		host, port, err = decodeAddrPort(br)
		if err != nil {
			return nil, err
		}
	case CmdMux:
		// Mux 没有后续目标，这里最小实现直接返回不支持
		return nil, errors.New("VLESS Mux not supported in minimal server")
	default:
		return nil, fmt.Errorf("unknown cmd: 0x%x", cmd)
	}

	return &Request{
		Version:    v,
		UserID:     uid,
		AddonsRaw:  addons,
		Command:    cmd,
		TargetAddr: host,
		TargetPort: port,
		Flow:       flow,
	}, nil
}

// 最小响应头：回显同版本 + 0 长度 Addons
func WriteResponseHeader(w io.Writer, ver byte) error {
	_, err := w.Write([]byte{ver, 0x00})
	return err
}
