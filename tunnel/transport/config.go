package transport

import (
	"github.com/ninepeach/n4fd/config"
)

type Config struct {
	LocalHost       string                `json:"local_addr" yaml:"local-addr"`
	LocalPort       int                   `json:"local_port" yaml:"local-port"`
	RemoteHost      string                `json:"remote_addr" yaml:"remote-addr"`
	RemotePort      int                   `json:"remote_port" yaml:"remote-port"`
	TCP          TCPConfig          `json:"tcp" yaml:"tcp"`
	TransportPlugin TransportPluginConfig `json:"transport_plugin" yaml:"transport-plugin"`
}

type TransportPluginConfig struct {
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Type    string   `json:"type" yaml:"type"`
	Option  string   `json:"option" yaml:"option"`
	Arg     []string `json:"arg" yaml:"arg"`
}

type TCPConfig struct {
	PreferIPV4 bool `json:"prefer_ipv4" yaml:"prefer-ipv4"`
	KeepAlive  bool `json:"keep_alive" yaml:"keep-alive"`
	NoDelay    bool `json:"no_delay" yaml:"no-delay"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
