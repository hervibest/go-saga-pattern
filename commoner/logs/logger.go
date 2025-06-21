package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log interface {
	Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry
	Core() zapcore.Core
	DPanic(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Level() zapcore.Level
	Log(lvl zapcore.Level, msg string, fields ...zap.Field)
	Name() string
	Named(s string) *zap.Logger
	Panic(msg string, fields ...zap.Field)
	Sugar() *zap.SugaredLogger
	Sync() error
	Warn(msg string, fields ...zap.Field)
	With(fields ...zap.Field) *zap.Logger
	WithLazy(fields ...zap.Field) *zap.Logger
	WithOptions(opts ...zap.Option) *zap.Logger
}

func NewLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	// Menonaktifkan stack trace untuk level Error
	config.DisableStacktrace = true

	// Atau jika Anda ingin lebih spesifik:
	// config.EncoderConfig.StacktraceKey = "" // Menghapus stack trace sepenuhnya

	logger, err := config.Build(
		// Opsi tambahan untuk menghilangkan caller (opsional)
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}
