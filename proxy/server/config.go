package server

import (
    "github.com/ninepeach/n4fd/config"
)

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
	})
}
