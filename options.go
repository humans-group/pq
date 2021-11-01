package pq

import "go.uber.org/zap"

type Option func(o options) options

type options struct {
	Logger *zap.Logger
}

func WithLogger(logger *zap.Logger) Option {
	return func(o options) options {
		o.Logger = logger
		return o
	}
}
