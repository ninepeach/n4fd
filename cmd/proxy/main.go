package main

import (
	"encoding/base64"
	"flag"
	"log"
	"net"
	"strings"

	"github.com/ninepeach/n4fd/internal/inbound/reality"
	"github.com/ninepeach/n4fd/internal/util/uuid"
	"github.com/ninepeach/n4fd/internal/vless"
)

// 小适配器：让 vless.Handler 满足 reality.NextHandler
type vlessNext struct{ h *vless.Handler }

func (n vlessNext) HandleConn(c net.Conn) error { return n.h.HandleConn(c) }

func parsePrivBase64URLMust(s string) []byte {
	r := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), "-", "+"), "_", "/")
	for len(r)%4 != 0 {
		r += "="
	}
	b, err := base64.StdEncoding.DecodeString(r)
	if err != nil {
		log.Fatalf("decode priv: %v", err)
	}
	if len(b) != 32 {
		log.Fatalf("invalid priv len=%d, expect 32", len(b))
	}
	return b
}

func mustParseUUIDs(csv string) []uuid.UUID {
	if strings.TrimSpace(csv) == "" {
		log.Fatal("missing -uuids")
	}
	parts := strings.Split(csv, ",")
	out := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		u, err := uuid.Parse(strings.TrimSpace(p))
		if err != nil {
			log.Fatalf("bad uuid %q: %v", p, err)
		}
		out = append(out, u)
	}
	return out
}

func main() {
	var (
		listen = flag.String("listen", ":8443", "REALITY listen addr")
		priv   = flag.String("priv", "", "REALITY x25519 private key (base64url, 32B)")
		dest   = flag.String("dest", "www.cloudflare.com:443", "fallback dest (TLS1.3 + h2)")
		sniStr = flag.String("sni", "www.cloudflare.com", "allowed SNI list (comma-separated) or *")
		short  = flag.String("short", "", "short-id hex (16 chars) — optional")
		uuids  = flag.String("uuids", "", "VLESS UUIDs (comma-separated)")
		debug  = flag.Bool("debug", true, "enable verbose logs")
	)
	flag.Parse()
	_ = short // 占位，后续如需接入 short-id 校验再用

	if *priv == "" {
		log.Fatal("missing -priv")
	}

	// 1) VLESS 处理器（支持 Vision）
	vh := vless.NewHandler(mustParseUUIDs(*uuids), *debug)

	// 2) 直接构造 reality.Inbound（字段名与当前仓库保持一致；以下是最常见的一套）
	inb := &reality.Inbound{
		Listen:  *listen,
		PrivKey: parsePrivBase64URLMust(*priv),
		Dest:    *dest,
		Debug:   *debug,
		Next:    vlessNext{h: vh},
	}

	// SNI 白名单（若 reality.Inbound 暴露了 SNIAllow 字段就用；没有就忽略）
	if *sniStr != "*" {
		inb.SNIAllow = map[string]bool{}
		for _, n := range strings.Split(*sniStr, ",") {
			n = strings.TrimSpace(n)
			if n != "" {
				inb.SNIAllow[n] = true
			}
		}
	}

	ln, err := net.Listen("tcp", inb.Listen)
	if err != nil {
		log.Fatalf("listen %s: %v", inb.Listen, err)
	}
	log.Printf("Listening REALITY on %s  dest=%s", inb.Listen, inb.Dest)
	log.Fatal(inb.Serve(ln))
}
