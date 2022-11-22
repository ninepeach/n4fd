package redirector

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ninepeach/n4fd/common"
)

func TestRedirector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	redir := NewRedirector(ctx)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
        fmt.Println(err)
        panic(err)
	}

	conn1, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
        fmt.Println(err)
        panic(err)
	}

	conn2, err := l.Accept()
	if err != nil {
        fmt.Println(err)
        panic(err)
	}

	redirAddr, err := net.ResolveTCPAddr("tcp", "www.baidu.com:80")
	if err != nil {
        fmt.Println(err)
        panic(err)
	}
	
	redir.Redirect(&Redirection{
		Dial:        nil,
		RedirectTo:  redirAddr,
		InboundConn: conn2,
	})
	time.Sleep(time.Second)


	req, err := http.NewRequest("GET", "https://www.baidu.com/", nil)
	common.Must(err)
	req.Write(conn1)
	buf := make([]byte, 1024)
	conn1.Read(buf)
	fmt.Println(string(buf))
	if !strings.HasPrefix(string(buf), "HTTP/1.1 200 OK") {
		t.Fail()
	}
	cancel()
	conn1.Close()
	conn2.Close()
}
