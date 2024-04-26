package tlshs

import "fmt"

type (
	Version uint16 // TLS Record Version, also handshake version
)

// String method to return string of TLS version
func (v Version) String() string {
	if name, ok := VersionReg[v]; ok {
		return name
	}
	return fmt.Sprintf("%#v (unknown)", v)
}

const (
	VersionSSL30 Version = 0x0300
	VersionTLS10 Version = 0x0301
	VersionTLS11 Version = 0x0302
	VersionTLS12 Version = 0x0303
	VersionTLS13 Version = 0x0304
)

var VersionReg = map[Version]string{
	VersionSSL30: "SSL 3.0",
	VersionTLS10: "TLS 1.0",
	VersionTLS11: "TLS 1.1",
	VersionTLS12: "TLS 1.2",
	VersionTLS13: "TLS 1.3",
}
