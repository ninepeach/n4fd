package proxy

import (
	"flag"

	"github.com/ninepeach/n4fd/log"
	"github.com/ninepeach/n4fd/option"
)

type Option struct {
	path *string
}

func (o *Option) Name() string {
	return Name
}

func (o *Option) Handle() error {

	var data []byte
	var err error

	data = []byte(`
run-type: server
local-addr: 0.0.0.0
local-port: 13131
remote-addr: www.lanzou.com.lanzoujj.xyz
remote-port: 12023
shadowsocks:
  enabled: true
  method: AES-128-GCM
  password: haMLMXirByn6rGVh
`)
	proxy, err := NewProxyFromConfigData(data)

	if err != nil {

		log.Fatal(err.Error())
	}
	err = proxy.Run()
	if err != nil {
		log.Fatal(err.Error())
	}

	return nil
}

func (o *Option) Priority() int {
	return -1
}

func init() {
	option.RegisterHandler(&Option{
		path: flag.String("config", "config.yaml", "config filename (.yaml/.yml)"),
	})
}
