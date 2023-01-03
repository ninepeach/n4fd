package proxy

import (
    "bufio"
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "runtime"

    log "github.com/ninepeach/go-clog"

    "github.com/ninepeach/n4fd/constant"
    "github.com/ninepeach/n4fd/option"
)

func init() {
    option.RegisterHandler(&Option{
        path: flag.String("config", "", "n4fd config filename (.yaml/.yml)"),
    })
    option.RegisterHandler(&StdinOption{
        format: flag.String("stdin-format", "disabled", "Read from standard input yaml"),
    })
}

type Option struct {
    path *string
}

func (o *Option) Name() string {
    return Name
}

func (o *Option) Priority() int {
    return 1
}

func readConfig(file string) ([]byte, error) {

    defaultConfigPath := []string{
        "config.yml",
        "config.yaml",
    }

    var data []byte
    var err error

    switch file {
    case "":
        log.Info("no specified config file, use default path to detect config file")
        for _, file := range defaultConfigPath {
            log.Info("try to load config from default path: %s", file)
            data, err = ioutil.ReadFile(file)
            if err != nil {
                log.Info(err.Error())
                continue
            }
            break
        }
    default:
        data, err = ioutil.ReadFile(file)
        if err != nil {
            log.Info(err.Error())
        }
    }
    return data, err
}

func (o *Option) Handle() error {

    var data []byte
    var err error

    data, err = readConfig(*o.path)
    if (data == nil || err != nil) {
        log.Fatal("no valid config")
    }

    log.Info("n4fd %s initializing", constant.Version)
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


type StdinOption struct {
    format       *string
}

func (o *StdinOption) Name() string {
    return Name + "_STDIN"
}

func (o *StdinOption) Priority() int {
    return 0
}

func (o *StdinOption) Handle() error {

    fmt.Printf("n4fd %s (%s/%s)\n", constant.Version, runtime.GOOS, runtime.GOARCH)
    fmt.Println("Reading YAML configuration from stdin.")

    data, e := ioutil.ReadAll(bufio.NewReader(os.Stdin))
    if e != nil {
        log.Fatal("Failed to read from stdin: %s", e.Error())
    }

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
