package server

import (
    "context"

    "github.com/ninepeach/n4fd/config"
    "github.com/ninepeach/n4fd/proxy"
    "github.com/ninepeach/n4fd/tunnel/freedom"
    "github.com/ninepeach/n4fd/tunnel/ss"
    "github.com/ninepeach/n4fd/tunnel/transport"
)

const Name = "SERVER"

func init() {
    proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
        cfg := config.FromContext(ctx, Name).(*client.Config)
        ctx, cancel := context.WithCancel(ctx)
        transportServer, err := transport.NewServer(ctx, nil)
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
        fmt.Println(clientStack)

root := &proxy.Node{
            Name:       transport.Name,
            Next:       make(map[string]*proxy.Node),
            IsEndpoint: false,
            Context:    ctx,
            Server:     transportServer,
        }

        root = root.BuildNext(tls.Name)
        wsSubTree := root.BuildNext(websocket.Name)
        wsSubTree = wsSubTree.BuildNext(ss.Name)
        wsSubTree.BuildNext(socks.Name).IsEndpoint = true

        serverList := proxy.FindAllEndpoints(root)
        clientList, err := proxy.CreateClientStack(ctx, clientStack)
        if err != nil {
            cancel()
            return nil, err
        }
        return proxy.NewProxy(ctx, cancel, serverList, clientList), nil
    })
}
