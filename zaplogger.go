package hs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(level string, outputPaths, errorPaths []string) (*zap.Logger, error) {
	l := new(zapcore.Level)
	if err := l.Set(level); err != nil {
		return nil, err
	}
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(*l),
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorPaths,
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	}

	return cfg.Build()
}
