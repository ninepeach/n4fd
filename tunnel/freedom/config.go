package freedom

import "github.com/ninepeach/n4fd/config"

type Config struct {
	LocalHost string    `json:"local_addr" yaml:"local-addr"`
	LocalPort int       `json:"local_port" yaml:"local-port"`
	TCP       TCPConfig `json:"tcp" yaml:"tcp"`
}

type TCPConfig struct {
	PreferIPV4 bool `json:"prefer_ipv4" yaml:"prefer-ipv4"`
	KeepAlive  bool `json:"keep_alive" yaml:"keep-alive"`
	NoDelay    bool `json:"no_delay" yaml:"no-delay"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			TCP: TCPConfig{
				PreferIPV4: false,
				NoDelay:    true,
				KeepAlive:  true,
			},
		}
	})
}
