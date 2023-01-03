package main

import (
    "flag"

    _ "github.com/ninepeach/n4fd/component"
    "github.com/ninepeach/n4fd/option"
    log "github.com/ninepeach/go-clog"

)

func main() {
    flag.Parse()
    for {
        h, err := option.PopOptionHandler()
        if err != nil {
            log.Fatal("invalid options %s", err.Error())
        }

        err = h.Handle()
        if err == nil {
            break
        }
    }
}
