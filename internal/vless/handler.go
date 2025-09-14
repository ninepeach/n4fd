package vless

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/ninepeach/n4fd/internal/util/uuid"
)

type Handler struct {
	// 允许的 UUID 集（白名单）
	Users map[uuid.UUID]struct{}

	// 出站拨号（可被替换）；默认 direct
	DialContext func(ctx context.Context, network, address string) (net.Conn, error)

	// 日志开关
	Debug bool
}

func NewHandler(uuids []uuid.UUID, debug bool) *Handler {
	m := make(map[uuid.UUID]struct{}, len(uuids))
	for _, u := range uuids {
		m[u] = struct{}{}
	}
	return &Handler{
		Users: m,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, address)
		},
		Debug: debug,
	}
}

func (h *Handler) allow(u uuid.UUID) bool {
	_, ok := h.Users[u]
	return ok
}

func (h *Handler) HandleConn(c net.Conn) error {
	defer c.Close()

	br := bufio.NewReader(c)

	// 解析 VLESS 请求
	req, err := ParseRequest(br, h.allow)
	if err != nil {
		return fmt.Errorf("parse vless request: %w", err)
	}

	if h.Debug {
		fmt.Printf("VLESS: user=%x cmd=%d dst=%s:%d flow=%s\n", req.UserID, req.Command, req.TargetAddr, req.TargetPort, req.Flow)
	}

	// 只支持 TCP
	if req.Command != CmdTCP {
		return errors.New("only TCP is supported in minimal server")
	}

	// 回写响应头
	if err := WriteResponseHeader(c, req.Version); err != nil {
		return fmt.Errorf("write vless response header: %w", err)
	}

	// 拨号到目标
	dst := net.JoinHostPort(req.TargetAddr, fmt.Sprintf("%d", req.TargetPort))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rc, err := h.DialContext(ctx, "tcp", dst)
	if err != nil {
		return fmt.Errorf("dial %s: %w", dst, err)
	}
	defer rc.Close()

	// 把 reader 中还残留的字节（如版本后续已经读取一部分）交给 rc
	// 这里 br 里此时只剩 body；我们需要把 br 连续读 => rc，反之亦然
	errc := make(chan error, 2)

	go func() {
		// client->remote
		_, e := io.Copy(rc, br)
		_ = rc.(*net.TCPConn).CloseWrite()
		errc <- e
	}()
	go func() {
		// remote->client
		_, e := io.Copy(c, rc)
		_ = c.(*net.TCPConn).CloseWrite()
		errc <- e
	}()

	// 等任一方向出错/结束
	if h.Debug {
		if e := <-errc; e != nil && !errors.Is(e, io.EOF) {
			fmt.Println("pipe err:", e)
		}
		<-errc
		return nil
	}
	<-errc
	<-errc
	return nil
}
