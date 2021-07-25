package zap

import (
	"io"

	"github.com/zhiyunliu/velocity/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	logger.Options
}

type callerSkipKey struct{}

func WithCallerSkip(i int) logger.Option {
	return logger.SetOption(callerSkipKey{}, i)
}

type configKey struct{}

// WithConfig pass zap.Config to logger
func WithConfig(c zap.Config) logger.Option {
	return logger.SetOption(configKey{}, c)
}

type encoderConfigKey struct{}

// WithEncoderConfig pass zapcore.EncoderConfig to logger
func WithEncoderConfig(c zapcore.EncoderConfig) logger.Option {
	return logger.SetOption(encoderConfigKey{}, c)
}

type namespaceKey struct{}

func WithNamespace(namespace string) logger.Option {
	return logger.SetOption(namespaceKey{}, namespace)
}

type writerKey struct{}

func WithOutput(out io.Writer) logger.Option {
	return logger.SetOption(writerKey{}, out)
}
