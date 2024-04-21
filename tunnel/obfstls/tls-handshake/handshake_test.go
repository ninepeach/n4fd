package tlshs

import (
	"context"
	"fmt"
	"testing"

	"github.com/ninepeach/n4fd/common"
	"github.com/ninepeach/n4fd/config"
	"github.com/ninepeach/n4fd/log"
	_ "github.com/ninepeach/n4fd/log/golog"
	"github.com/ninepeach/n4fd/tunnel/transport"
)

func TestTransport(t *testing.T) {

	serverCfg := &transport.Config{
		LocalHost:  "0.0.0.0",
		LocalPort:  13131,
		RemoteHost: "127.0.0.1",
		RemotePort: common.PickPort("tcp", "127.0.0.1"),
	}
	sctx := config.WithConfig(context.Background(), transport.Name, serverCfg)

	s, err := transport.NewServer(sctx, nil)
	common.Must(err)

	for {
		conn, err := s.AcceptConn(nil)
		fmt.Println("Accept", conn, err)
		common.Must(err)

		err = serverHandshake(conn)
		fmt.Println(err)
	}
	s.Close()
}

func init() {
	log.SetLogLevel(log.LogLevel(0))
}
