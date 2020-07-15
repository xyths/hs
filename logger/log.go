package logger

import (
	"go.uber.org/zap"
	"os"
)

var Sugar *zap.SugaredLogger

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
	Sugar = logger.Sugar()
}
