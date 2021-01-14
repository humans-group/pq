package pq

import (
	"context"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerAdapter struct {
	logger *zap.Logger
}

func newLoggerAdapter(logger *zap.Logger) *loggerAdapter {
	return &loggerAdapter{logger: logger.WithOptions(zap.AddCallerSkip(1))}
}

func (a *loggerAdapter) Log(ctx context.Context, pgxLevel pgx.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]zapcore.Field, len(data)+1)
	i := 0
	for k, v := range data {
		fields[i] = zap.Reflect(k, v)
		i++
	}
	fields[i] =  zap.Stringer("PGX_LOG_LEVEL", pgxLevel)

	level := a.level(pgxLevel)

	a.logger.Check(level, msg).Write(fields...)
}

func (a *loggerAdapter) level(level pgx.LogLevel) zapcore.Level {
	switch level {
	case pgx.LogLevelTrace, pgx.LogLevelDebug, pgx.LogLevelInfo:
		return zapcore.DebugLevel
	case pgx.LogLevelWarn:
		return zapcore.WarnLevel
	case pgx.LogLevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.ErrorLevel
	}
}