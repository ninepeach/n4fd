package golog

import (
	"github.com/ninepeach/n4fd/log"
	"io"
	glog "log"
	"os"
)

func init() {
	log.RegisterLogger(&GoLogger{})
}

type GoLogger struct {
	logLevel log.LogLevel
}

func (l *GoLogger) SetLogLevel(level log.LogLevel) {
	l.logLevel = level
}

func (l *GoLogger) Fatal(v ...interface{}) {
	if l.logLevel <= log.FatalLevel {
		glog.Fatal(v...)
	}
	os.Exit(1)
}

func (l *GoLogger) Fatalf(format string, v ...interface{}) {
	if l.logLevel <= log.FatalLevel {
		glog.Fatalf(format, v...)
	}
	os.Exit(1)
}

func (l *GoLogger) Error(v ...interface{}) {
	if l.logLevel <= log.ErrorLevel {
		glog.Println(v...)
	}
}

func (l *GoLogger) Errorf(format string, v ...interface{}) {
	if l.logLevel <= log.ErrorLevel {
		glog.Printf(format, v...)
	}
}

func (l *GoLogger) Warn(v ...interface{}) {
	if l.logLevel <= log.WarnLevel {
		glog.Println(v...)
	}
}

func (l *GoLogger) Warnf(format string, v ...interface{}) {
	if l.logLevel <= log.WarnLevel {
		glog.Printf(format, v...)
	}
}

func (l *GoLogger) Info(v ...interface{}) {
	if l.logLevel <= log.InfoLevel {
		glog.Println(v...)
	}
}

func (l *GoLogger) Infof(format string, v ...interface{}) {
	if l.logLevel <= log.InfoLevel {
		glog.Printf(format, v...)
	}
}

func (l *GoLogger) Debug(v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		glog.Println(v...)
	}
}

func (l *GoLogger) Debugf(format string, v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		glog.Printf(format, v...)
	}
}

func (l *GoLogger) Trace(v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		glog.Println(v...)
	}
}

func (l *GoLogger) Tracef(format string, v ...interface{}) {
	if l.logLevel <= log.AllLevel {
		glog.Printf(format, v...)
	}
}

func (l *GoLogger) SetOutput(io.Writer) {
	// do nothing
}
