package golog

import (
	"errors"
	"github.com/ninepeach/n4fd/log"
	"testing"
)

func TestPrintSomething(t *testing.T) {

	log.SetLogLevel(log.ErrorLevel)
	log.Info("hello info")
	log.Warn("hello warn")
	log.Debug("hello debug")
	log.Error(errors.New("Sample Error"))
}
