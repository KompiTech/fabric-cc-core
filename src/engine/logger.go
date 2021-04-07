package engine

import (
	"github.com/op/go-logging"
)

type Logger struct {
	lgr *logging.Logger
}

func NewLogger() Logger {
	lgr := logging.MustGetLogger("cc-core")
	be := logging.SetBackend()
	be.SetLevel(logging.ERROR, "")

	lgr.SetBackend(be)

	return Logger{
		lgr,
	}
}

func (l Logger) Warningf(format string, args ...interface{}) {
	l.lgr.Warningf(format, args...)
}

func (l Logger) Debugf(format string, args ...interface{}) {
	l.lgr.Debugf(format, args...)
}

func (l Logger) Errorf(format string, args ...interface{}) {
	l.lgr.Errorf(format, args...)
}

func (l Logger) Infof(format string, args ...interface{}) {
	l.lgr.Infof(format, args...)
}
