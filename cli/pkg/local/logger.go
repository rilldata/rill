package local

import (
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(verbose bool) *zap.Logger {
	lvl := zap.InfoLevel
	if verbose {
		lvl = zap.DebugLevel
	}
	return newBaseLogger().WithOptions(zap.IncreaseLevel(lvl))
}

func NewServerLogger(verbose bool) *zap.Logger {
	lvl := zap.ErrorLevel
	if verbose {
		lvl = zap.DebugLevel
	}
	return newBaseLogger().WithOptions(zap.IncreaseLevel(lvl))
}

func newBaseLogger() *zap.Logger {
	conf := zap.NewDevelopmentEncoderConfig()
	conf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(conf),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	))
}
