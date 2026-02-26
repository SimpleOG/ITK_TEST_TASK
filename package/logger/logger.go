package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}

type logger struct {
	logger *zap.Logger
}

func NewLogger(level zapcore.Level) (Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	log, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &logger{logger: log}, nil
}
func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}
func (l *logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}
func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}
func (l *logger) Error(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}
func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}
