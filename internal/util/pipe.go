package util

import (
	"io"
	"net"
	"sync"
)

type closeWriter interface{ CloseWrite() error }

func halfClose(c net.Conn) {
	if cw, ok := c.(closeWriter); ok {
		_ = cw.CloseWrite()
	}
}

func BidiCopy(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(b, a)
		halfClose(b)
	}()
	go func() {
		defer wg.Done()
		io.Copy(a, b)
		halfClose(a)
	}()
	wg.Wait()
}
