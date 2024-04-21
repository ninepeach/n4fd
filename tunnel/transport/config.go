package transport

import (
	"github.com/ninepeach/n4fd/config"
)

type Config struct {
	LocalHost  string `json:"addr" yaml:"addr"`
	LocalPort  int    `json:"port" yaml:"port"`
	RemoteHost string `json:"remote_addr" yaml:"remote-addr"`
	RemotePort int    `json:"remote_port" yaml:"remote-port"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			LocalHost: "0.0.0.0",
			LocalPort: 13131,
		}
	})
}
