package log

import (
	"go.uber.org/zap"
	"os"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

func init() {
	env := os.Getenv("GRID_ENV")
	switch env {
	case "PRODUCTION":
		Logger, _ = zap.NewProduction()
	case "DEVELOPMENT":
		Logger, _ = zap.NewDevelopment()
	default:
		Logger, _ = zap.NewDevelopment()
	}
	Sugar = Logger.Sugar()
}

func Debug(args ...interface{}) {
	Sugar.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	Sugar.Debugf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	Sugar.Debugw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	Sugar.Info(args...)
}

func Infof(template string, args ...interface{}) {
	Sugar.Infof(template, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	Sugar.Infow(msg, keysAndValues...)
}

func Warn(args ...interface{}) {
	Sugar.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	Sugar.Warnf(template, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	Sugar.Warnw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	Sugar.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	Sugar.Errorf(template, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	Sugar.Errorw(msg, keysAndValues...)
}

func DPanic(args ...interface{}) {
	Sugar.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	Sugar.DPanicf(template, args...)
}

func DPanicw(msg string, keysAndValues ...interface{}) {
	Sugar.DPanicw(msg, keysAndValues...)
}

func Panic(args ...interface{}) {
	Sugar.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	Sugar.Panicf(template, args...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	Sugar.Panicw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	Sugar.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	Sugar.Fatalf(template, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	Sugar.Fatalw(msg, keysAndValues...)
}
