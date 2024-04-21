package version

import (
	"errors"
	"flag"
	"fmt"
	"runtime"

	"github.com/ninepeach/n4fd/constant"
	"github.com/ninepeach/n4fd/option"
)

type versionOption struct {
	flag *bool
}

func (*versionOption) Name() string {
	return "version"
}

func (*versionOption) Priority() int {
	return 10
}

func (c *versionOption) Handle() error {
	if *c.flag {
		fmt.Println("N4FD", constant.Version)
		fmt.Println("Go Version:", runtime.Version())
		fmt.Println("Build OS/Arch:", runtime.GOOS+"/"+runtime.GOARCH)
		fmt.Println("Build Revision:", constant.BuildRevision)
		fmt.Println("Build Branch:", constant.BuildBranch)
		fmt.Println("Build Time:", constant.BuildTime)
		fmt.Println("")
		return nil
	}

	return errors.New("no set")
}

func init() {
	option.RegisterHandler(&versionOption{
		flag: flag.Bool("version", false, "Display version and help info"),
	})
}
