package obfs

import (
    "bytes"
    "io"
    "sync"

    "crypto/rand"
    "crypto/sha1"
    "encoding/base64"

    "github.com/ninepeach/n4fd/tunnel"
    log "github.com/ninepeach/go-clog"
)

const (
    OBFS_NEED_MORE = -1
    OBFS_ERROR = -2
    OBFS_OK = 0
)

var (
    Debug = true

    // DefaultUserAgent is the default HTTP User-Agent header used by HTTP and websocket.
    DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"
)

var keyGUID = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

func computeAcceptKey(challengeKey string) string {
    h := sha1.New()
    h.Write([]byte(challengeKey))
    h.Write(keyGUID)
    return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func generateChallengeKey() (string, error) {
    p := make([]byte, 16)
    if _, err := io.ReadFull(rand.Reader, p); err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(p), nil
}

type obfsHTTPConn struct {
    tunnel.Conn
    host           string
    rbuf           bytes.Buffer
    wbuf           bytes.Buffer
    isServer       bool
    headerDrained  bool
    handshaked     bool
    handshakeMutex sync.Mutex
}


func (c *obfsHTTPConn) Read(b []byte) (n int, err error) {
    return c.Conn.Read(b)
}

func (c *obfsHTTPConn) drainHeader(b []byte) (err error) {
    if c.headerDrained {
        return
    }

    pos := bytes.Index(b, []byte("\r\n\r\n"))
    log.Debug("pos :%d", pos)
    if pos == -1 {
        return
    }

    part1 := b[:pos]
    part2 := b[pos+2:]
    log.Debug("pos part1:%s", string(part1))
    log.Debug("pos part2:%s", string(part2))

    c.headerDrained = true

    return nil
}

func (c *obfsHTTPConn) Write(b []byte) (n int, err error) {
    log.Debug("fucking cao write 1")

    if c.isServer {
        if c.headerDrained {
            return
        }

        pos := bytes.Index(b, []byte("\r\n\r\n"))
        log.Debug("pos :%d", pos)
        if pos == -1 {
            return
        }

        part1 := b[:pos]
        part2 := b[pos+2:]
        log.Debug("pos part1:%s", string(part1))
        log.Debug("pos part2:%s", string(part2))

        b = part2
        c.headerDrained = true
    }
    
    return c.Conn.Write(b)
}

