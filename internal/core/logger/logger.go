package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.StacktraceKey = ""
	cfg.Level = lvl
	cfg.OutputPaths = []string{"stdout", "./app.log"}
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	L = zl
	return nil
}
