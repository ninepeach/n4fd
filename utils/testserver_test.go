package utils

import (
    "fmt"
    "net"
    "strings"
    "net/http"
    "testing"

    "github.com/ninepeach/n4fd/common"
)

func TestTestserver(t *testing.T) {
    runHelloHTTPServer("127.0.0.1:8081")

    conn, err := net.Dial("tcp", "127.0.0.1:8081")
    if err != nil {
        fmt.Println(err)
        panic(err)
    }

    req, err := http.NewRequest("GET", "http://127.0.0.1:8081", nil)
    common.Must(err)
    req.Write(conn)
    buf := make([]byte, 1024)
    conn.Read(buf)
    //fmt.Println(string(buf))
    if !strings.Contains(string(buf), "HelloWorld") {
        t.Fail()
    }
}

