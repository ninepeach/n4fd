# ss example

server:
```
package main

import (
    "fmt"
    "net"
    "os"
    "github.com/ninepeach/n4fd/ss/core"
)

const (
    CONN_HOST = "localhost"
    CONN_PORT = "3333"
    CONN_TYPE = "tcp"
)

func main() {
    cipher, err := core.PickCipher("AES-128-GCM", nil, "4fd")
    if err != nil {
        fmt.Println("Error init cipher", err.Error())
        os.Exit(1)
    }
    l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    defer l.Close()
    fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

    for {
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        conn = cipher.StreamConn(conn)

        go handleRequest(conn)
    }
}

func handleRequest(conn net.Conn) {
    buf := [20]byte{}
    _, err := conn.Read(buf[:])
    if err != nil {
        fmt.Println("read failed:", err)
        return
    }
    fmt.Println("Got: ", string(buf[:]))

    _, err = conn.Write([]byte("Message received.\n"))
    if err != nil {
        fmt.Println("write failed:", err)
        return
    }
    conn.Close()
}
```

client:
```
package main

import (
    "fmt"
    "net"
    "os"
    "github.com/ninepeach/go-libss/core"
)

const (
    CONN_HOST = "localhost"
    CONN_PORT = "3333"
    CONN_TYPE = "tcp"
)

func main() {
    cipher, err := core.PickCipher("AES-128-GCM", nil, "4fd")

    conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    conn = cipher.StreamConn(conn)
    ReadNWrite(conn)
    conn.Close()
}

func ReadNWrite(conn net.Conn) {
    _, err := conn.Write([]byte("12345678\r\n"))
    if err != nil {
        fmt.Println("failed:", err)
        return
    }

    buf := [20]byte{}
    _, err = conn.Read(buf[:])
    if err != nil {
        fmt.Println("read failed:", err)
        return
    }
    fmt.Println("Got: ", string(buf[:]))

}

```
