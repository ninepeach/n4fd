package tlshs

const (
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 uint16 = 0xc02b
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 uint16 = 0xc02c
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384   uint16 = 0xc030
)

func ParseCipherSuites(raw []byte) ([]uint16, error) {
	ret := make([]uint16, 0, 3)

	for i := 0; i < len(raw); i += 2 {
		c := (uint16(raw[i]) << 8) + uint16(raw[i+1])
		switch c {
		case TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
			ret = append(ret, TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
		case TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
			ret = append(ret, TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384)
		case TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
			ret = append(ret, TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
		default:
		}
	}

	return ret, nil
}
