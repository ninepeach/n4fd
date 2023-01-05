package server

import (
    "fmt"
    "context"

    "github.com/ninepeach/n4fd/proxy"
    "github.com/ninepeach/n4fd/tunnel/transport"
    "github.com/ninepeach/n4fd/tunnel/freedom"
    "github.com/ninepeach/n4fd/tunnel/ss"
    "github.com/ninepeach/n4fd/tunnel/obfs"
)

const Name = "SERVER"

func init() {
    proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
        ctx, cancel := context.WithCancel(ctx)
        adapterServer, err := transport.NewServer(ctx, nil)
        if err != nil {
            cancel()
            return nil, err
        }
        clientStack := []string{freedom.Name}
        /*
        if cfg.Router.Enabled {
            clientStack = []string{freedom.Name, router.Name}
        }
        */
        //czw debug
        fmt.Println("czw debug", clientStack)

        root := &proxy.Node{
            Name:       transport.Name,
            Next:       make(map[string]*proxy.Node),
            IsEndpoint: false,
            Context:    ctx,
            Server:     adapterServer,
        }

        root.BuildNext(ss.Name).BuildNext(obfs.Name).IsEndpoint = true

        
        clientList, err := proxy.CreateClientStack(ctx, clientStack)
        if err != nil {
            cancel()
            return nil, err
        }
        serverList := proxy.FindAllEndpoints(root)
        return proxy.NewProxy(ctx, cancel, serverList, clientList), nil
    })
}
