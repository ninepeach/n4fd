package server

import (
    "fmt"
    "context"

    "github.com/ninepeach/n4fd/proxy"
    "github.com/ninepeach/n4fd/tunnel/transport"
    "github.com/ninepeach/n4fd/tunnel/adapter"
    "github.com/ninepeach/n4fd/tunnel/ss"
    "github.com/ninepeach/n4fd/tunnel/obfs"
    "github.com/ninepeach/n4fd/tunnel/http"
    "github.com/ninepeach/n4fd/tunnel/socks"
)

const Name = "SERVER"

func init() {
    proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
        ctx, cancel := context.WithCancel(ctx)
        adapterServer, err := adapter.NewServer(ctx, nil)
        if err != nil {
            cancel()
            return nil, err
        }
        clientStack := []string{transport.Name, ss.Name, obfs.Name}
        /*
        if cfg.Router.Enabled {
            clientStack = []string{freedom.Name, router.Name}
        }
        */
        //czw debug
        fmt.Println("czw debug", clientStack)

        root := &proxy.Node{
            Name:       adapter.Name,
            Next:       make(map[string]*proxy.Node),
            IsEndpoint: false,
            Context:    ctx,
            Server:     adapterServer,
        }

        root.BuildNext(http.Name).IsEndpoint = true
        root.BuildNext(socks.Name).IsEndpoint = true

        serverList := proxy.FindAllEndpoints(root)
        clientList, err := proxy.CreateClientStack(ctx, clientStack)
        if err != nil {
            cancel()
            return nil, err
        }
        return proxy.NewProxy(ctx, cancel, serverList, clientList), nil
    })
}
