package proxy

import "github.com/ninepeach/n4fd/config"

type Config struct {
	RunType  string `json:"run_type" yaml:"run-type"`
	LogLevel int    `json:"log_level" yaml:"log-level"`
	LogFile  string `json:"log_file" yaml:"log-file"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			RunType: "server",
			LogLevel: 1,
			LogFile: "n4fd.log",
		}
	})
}
