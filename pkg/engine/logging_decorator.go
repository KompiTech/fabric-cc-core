package engine

//import (
//	"fmt"
//	"runtime"
//	"strings"
//
//	"github.com/google/martian/log"
//)
//
//type LoggingDecorator struct{}
//
//func (l LoggingDecorator) Errorf(format string, args ...interface{}) {
//	log.Errorf(decorate(format), args...)
//}
//
//func (l LoggingDecorator) Infof(format string, args ...interface{}) {
//	log.Infof(decorate(format), args...)
//}
//
//func (l LoggingDecorator) Debugf(format string, args ...interface{}) {
//	log.Debugf(decorate(format), args...)
//}
//
//func NewLoggingDecorator() LoggingDecorator {
//	return LoggingDecorator{}
//}
//
//// decorate adds originating file and line to logging message
//func decorate(input string) string {
//	_, file, line, _ := runtime.Caller(2) // skip decorate() and wrapping func
//	lastSlash := strings.LastIndex(file, "/")
//
//	return fmt.Sprintf("%s:%d %s", file[lastSlash+1:], line, input)
//}
