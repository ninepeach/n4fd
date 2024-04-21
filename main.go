package main

import (
	"flag"
	_ "github.com/ninepeach/n4fd/component"

	"github.com/ninepeach/n4fd/common"
	"github.com/ninepeach/n4fd/log"
	"github.com/ninepeach/n4fd/option"
)

func main() {
	flag.Parse()
	for {
		h, err := option.PopOptionHandler()

		if err != nil {
			log.Fatal(common.NewError("invalid options").Base(err))
		}

		err = h.Handle()
		if err == nil {
			break
		}
	}
}
