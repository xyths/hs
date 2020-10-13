package hs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(conf LogConf) (*zap.Logger, error) {
	l := new(zapcore.Level)
	if err := l.Set(conf.Level); err != nil {
		return nil, err
	}
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(*l),
		OutputPaths:      conf.Outputs,
		ErrorOutputPaths: conf.Errors,
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	}

	return cfg.Build()
}
