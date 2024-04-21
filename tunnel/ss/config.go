package ss

import "github.com/ninepeach/n4fd/config"

type ShadowsocksConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Method   string `json:"method" yaml:"method"`
	Password string `json:"password" yaml:"password"`
}

type Config struct {
	Shadowsocks ShadowsocksConfig `json:"shadowsocks" yaml:"shadowsocks"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			Shadowsocks: ShadowsocksConfig{
				Method:   "AES-128-GCM",
				Password: "haMLMXirByn6rGVh",
			},
		}
	})
}
