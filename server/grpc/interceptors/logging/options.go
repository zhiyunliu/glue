/*
 * @Author: lwnmengjing
 * @Date: 2021/5/19 11:14 上午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/5/19 11:14 上午
 */

package logging

import (
	"context"
	"time"

	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"github.com/zhiyunliu/velocity/log"
	"github.com/zhiyunliu/velocity/server/grpc/interceptors/logging/ctxlog"
	"google.golang.org/grpc/codes"
)

var (
	defaultOptions = &options{
		levelFunc:       DefaultCodeToLevel,
		shouldLog:       grpc_logging.DefaultDeciderMethod,
		codeFunc:        grpc_logging.DefaultErrorToCode,
		durationFunc:    DefaultDurationToField,
		messageFunc:     DefaultMessageProducer,
		timestampFormat: time.RFC3339,
	}
)

type options struct {
	levelFunc       CodeToLevel
	shouldLog       grpc_logging.Decider
	codeFunc        grpc_logging.ErrorToCode
	durationFunc    DurationToField
	messageFunc     MessageProducer
	timestampFormat string
}

type Option func(*options)

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(code codes.Code) log.Level

// DurationToField function defines how to produce duration fields for logging
type DurationToField func(duration time.Duration) ctxlog.Fields

// WithDecider customizes the function for deciding if the gRPC interceptor logs should log.
func WithDecider(f grpc_logging.Decider) Option {
	return func(o *options) {
		o.shouldLog = f
	}
}

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithCodes customizes the function for mapping errors to error codes.
func WithCodes(f grpc_logging.ErrorToCode) Option {
	return func(o *options) {
		o.codeFunc = f
	}
}

// WithDurationField customizes the function for mapping request durations to Zap fields.
func WithDurationField(f DurationToField) Option {
	return func(o *options) {
		o.durationFunc = f
	}
}

// WithMessageProducer customizes the function for message formation.
func WithMessageProducer(f MessageProducer) Option {
	return func(o *options) {
		o.messageFunc = f
	}
}

// WithTimestampFormat customizes the timestamps emitted in the log fields.
func WithTimestampFormat(format string) Option {
	return func(o *options) {
		o.timestampFormat = format
	}
}

// MessageProducer produces a user defined log message
type MessageProducer func(ctx context.Context, msg string, level log.Level, code codes.Code, err error, duration ctxlog.Fields)

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateClientOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultClientCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

// DefaultCodeToLevel is the default implementation of gRPC return codes and interceptor log level for server side.
func DefaultCodeToLevel(code codes.Code) log.Level {
	switch code {
	case codes.OK:
		return log.InfoLevel
	case codes.Canceled:
		return log.InfoLevel
	case codes.Unknown:
		return log.ErrorLevel
	case codes.InvalidArgument:
		return log.InfoLevel
	case codes.DeadlineExceeded:
		return log.WarnLevel
	case codes.NotFound:
		return log.InfoLevel
	case codes.AlreadyExists:
		return log.InfoLevel
	case codes.PermissionDenied:
		return log.WarnLevel
	case codes.Unauthenticated:
		return log.InfoLevel // unauthenticated requests can happen
	case codes.ResourceExhausted:
		return log.WarnLevel
	case codes.FailedPrecondition:
		return log.WarnLevel
	case codes.Aborted:
		return log.WarnLevel
	case codes.OutOfRange:
		return log.WarnLevel
	case codes.Unimplemented:
		return log.ErrorLevel
	case codes.Internal:
		return log.ErrorLevel
	case codes.Unavailable:
		return log.WarnLevel
	case codes.DataLoss:
		return log.ErrorLevel
	default:
		return log.ErrorLevel
	}
}

// DefaultClientCodeToLevel is the default implementation of gRPC return codes to log levels for client side.
func DefaultClientCodeToLevel(code codes.Code) log.Level {
	switch code {
	case codes.OK:
		return log.DebugLevel
	case codes.Canceled:
		return log.DebugLevel
	case codes.Unknown:
		return log.InfoLevel
	case codes.InvalidArgument:
		return log.DebugLevel
	case codes.DeadlineExceeded:
		return log.InfoLevel
	case codes.NotFound:
		return log.DebugLevel
	case codes.AlreadyExists:
		return log.DebugLevel
	case codes.PermissionDenied:
		return log.InfoLevel
	case codes.Unauthenticated:
		return log.InfoLevel // unauthenticated requests can happen
	case codes.ResourceExhausted:
		return log.DebugLevel
	case codes.FailedPrecondition:
		return log.DebugLevel
	case codes.Aborted:
		return log.DebugLevel
	case codes.OutOfRange:
		return log.DebugLevel
	case codes.Unimplemented:
		return log.WarnLevel
	case codes.Internal:
		return log.WarnLevel
	case codes.Unavailable:
		return log.WarnLevel
	case codes.DataLoss:
		return log.WarnLevel
	default:
		return log.InfoLevel
	}
}

// DefaultDurationToField is the default implementation of converting request duration to a Zap field.
var DefaultDurationToField = DurationToTimeMillisField

// DurationToTimeMillisField converts the duration to milliseconds and uses the key `grpc.time_ms`.
func DurationToTimeMillisField(duration time.Duration) ctxlog.Fields {
	return *ctxlog.NewFields("grpc.time_ms", durationToMilliseconds(duration))
}

// DurationToDurationField uses a Duration field to log the request duration
// and leaves it up to Zap's encoder settings to determine how that is output.
func DurationToDurationField(duration time.Duration) map[string]interface{} {
	return map[string]interface{}{"grpc.duration": duration}
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}

// DefaultMessageProducer writes the default message
func DefaultMessageProducer(ctx context.Context, msg string, level log.Level, code codes.Code, err error, duration ctxlog.Fields) {
	// re-extract logger from newCtx, as it may have extra fields that changed in the holder.
	fields := duration
	fields.Set("grpc.code", code.String())
	ctxlog.Extract(ctx).WithFields(fields.Values()).Log(level, msg, err)
	//if err != nil {
	//	ctxlog.Extract(ctx).WithFields(fields.Values()).Error(msg, err)
	//	return
	//}
	//ctxlog.Extract(ctx).WithFields(fields.Values()).Info(msg)
}
