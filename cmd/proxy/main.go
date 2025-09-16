package main

import (
	"encoding/base64"
	"encoding/hex"
	"flag"
	"log"
	"net"
	"strings"
	"time"

	realityin "github.com/ninepeach/n4fd/internal/inbound/reality"
	vless "github.com/ninepeach/n4fd/internal/vless"
)

func mustPrivBase64URL(s string) []byte {
	// accept base64url or base64
	r := strings.NewReplacer("-", "+", "_", "/").Replace(strings.TrimSpace(s))
	for len(r)%4 != 0 {
		r += "="
	}
	key, err := base64.StdEncoding.DecodeString(r)
	if err != nil {
		log.Fatalf("decode private key failed: %v", err)
	}
	if len(key) != 32 {
		log.Fatalf("invalid private key length: %d (expect 32)", len(key))
	}
	return key
}

func parseShortMap(short string, debug bool) map[[8]byte]bool {
	m := make(map[[8]byte]bool)
	short = strings.TrimSpace(short)
	if short == "" {
		// allow empty short-id (many clients send empty by default)
		var z [8]byte
		m[z] = true
		if debug {
			log.Printf("DEBUG allow empty short-id")
		}
		return m
	}
	// accept one or multiple short-ids separated by comma
	parts := strings.Split(short, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			var z [8]byte
			m[z] = true
			continue
		}
		if len(p) != 16 {
			log.Fatalf("short-id must be 16 hex chars: %q", p)
		}
		raw, err := hex.DecodeString(p)
		if err != nil {
			log.Fatalf("short-id hex decode failed: %v", err)
		}
		var k [8]byte
		copy(k[:], raw)
		m[k] = true
	}
	return m
}

func parseUUIDList(uuids string) map[[16]byte]bool {
	m := make(map[[16]byte]bool)
	for _, u := range strings.Split(uuids, ",") {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		var raw [16]byte
		b, err := hex.DecodeString(strings.ReplaceAll(strings.ReplaceAll(u, "-", ""), " ", ""))
		if err != nil || len(b) != 16 {
			log.Fatalf("invalid uuid: %s", u)
		}
		copy(raw[:], b)
		m[raw] = true
	}
	return m
}

func main() {
	var (
		listen  = flag.String("listen", ":8443", "TCP listen address")
		privStr = flag.String("priv", "", "REALITY private key (base64url/base64, 32 bytes)")
		dest    = flag.String("dest", "", "REALITY dial dest 'host:443' (recommend an IP:443 to avoid Fake-IP loops)")
		sni     = flag.String("sni", "*", "allowed SNI list (comma). Use '*' to allow any")
		short   = flag.String("short", "", "short-id (16 hex) or comma-separated multiple, empty allowed")
		uuids   = flag.String("uuids", "", "allowed VLESS UUIDs (comma separated, hyphens ok)")
		debug   = flag.Bool("debug", true, "enable debug logs")
	)
	flag.Parse()

	if *privStr == "" {
		log.Fatal("missing -priv")
	}
	if *dest == "" {
		log.Fatal("missing -dest (strongly recommend an IP:443)")
	}
	priv := mustPrivBase64URL(*privStr)

	var sniAllow map[string]bool
	if strings.TrimSpace(*sni) != "*" {
		sniAllow = map[string]bool{}
		for _, name := range strings.Split(*sni, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				sniAllow[name] = true
			}
		}
	} // nil => allow any

	shortMap := parseShortMap(*short, *debug)
	uuidAllow := parseUUIDList(*uuids)
	if len(uuidAllow) == 0 {
		log.Fatal("no UUIDs provided; set -uuids")
	}

	// Build VLESS handler with UUID whitelist
	vh := &vless.Handler{
		Allow:     uuidAllow,
		Dialer:    &net.Dialer{Timeout: 15 * time.Second},
		EnableLog: *debug,
	}

	// REALITY inbound
	ri := &realityin.Inbound{
		Listen:   *listen,
		PrivKey:  priv,
		Dest:     *dest,
		SNIAllow: sniAllow,
		ShortIDs: shortMap,
		Debug:    *debug,
		Next:     vh, // hand off to VLESS on success
	}

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen %s: %v", *listen, err)
	}
	log.Printf("INFO listening on %s (dest=%s sni=%v short-ids=%d)", *listen, *dest, func() any {
		if sniAllow == nil {
			return "*"
		}
		return sniAllow
	}(), len(shortMap))

	if err := ri.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
