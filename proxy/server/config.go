package server

import (
    "github.com/ninepeach/n4fd/config"
)

type ShadowsocksConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type TransportPluginConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type Config struct {
	Shadowsocks     ShadowsocksConfig     `json:"shadowsocks" yaml:"shadowsocks"`
	TransportPlugin TransportPluginConfig `json:"transport_plugin" yaml:"transport-plugin"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
