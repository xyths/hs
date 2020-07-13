package logger

import (
	"go.uber.org/zap"
	"os"
)

var sugar *zap.SugaredLogger

func init() {
	var logger *zap.Logger
	env := os.Getenv("GRID_ENV")
	switch env {
	case "PRODUCTION":
		logger, _ = zap.NewProduction()
	case "DEVELOPMENT":
		logger, _ = zap.NewDevelopment()
	default:
		logger, _ = zap.NewDevelopment()
	}
	sugar = logger.Sugar()
}

func Debug(args ...interface{}) {
	sugar.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	sugar.Debugw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	sugar.Info(args...)
}

func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	sugar.Infow(msg, keysAndValues...)
}

func Warn(args ...interface{}) {
	sugar.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	sugar.Warnw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	sugar.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	sugar.Errorw(msg, keysAndValues...)
}

func DPanic(args ...interface{}) {
	sugar.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	sugar.DPanicf(template, args...)
}

func DPanicw(msg string, keysAndValues ...interface{}) {
	sugar.DPanicw(msg, keysAndValues...)
}

func Panic(args ...interface{}) {
	sugar.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	sugar.Panicf(template, args...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	sugar.Panicw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	sugar.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	sugar.Fatalw(msg, keysAndValues...)
}
