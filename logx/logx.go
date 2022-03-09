package logx

import (
	"github.com/sirupsen/logrus"
	"runtime"
)

func Error(args ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	logrus.WithFields(logrus.Fields{"func": runtime.FuncForPC(pc).Name()}).Errorln(args)
}
func Warn(args ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	logrus.WithFields(logrus.Fields{"func": runtime.FuncForPC(pc).Name()}).Warnln(args)
}
func Info(args ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	logrus.WithFields(logrus.Fields{"func": runtime.FuncForPC(pc).Name()}).Infoln(args)
}
func Debug(args ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	logrus.WithFields(logrus.Fields{"func": runtime.FuncForPC(pc).Name()}).Debugln(args)
}
func Trace(args ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	logrus.WithFields(logrus.Fields{"func": runtime.FuncForPC(pc).Name()}).Traceln(args)
}
