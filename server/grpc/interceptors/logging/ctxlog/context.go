package ctxlog

import (
	"context"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/zhiyunliu/velocity/log"
)

type ctxMarker struct{}

type ctxLogger struct {
	logger *log.Wrapper
	fields map[string]interface{}
}

var (
	ctxMarkerKey = &ctxMarker{}
	nullLogger   = log.NewWrapper(log.DefaultLogger)
)

// AddFields adds logger fields to the log.
func AddFields(ctx context.Context, fields map[string]interface{}) {
	l, ok := ctx.Value(ctxMarkerKey).(*ctxLogger)
	if !ok || l == nil {
		return
	}
	for k, v := range fields {
		l.fields[k] = v
	}
}

// Extract takes the call-scoped Logger from grpc_logger middleware.
//
// It always returns a Logger that has all the grpc_ctxtags updated.
func Extract(ctx context.Context) *log.Wrapper {
	l, ok := ctx.Value(ctxMarkerKey).(*ctxLogger)
	if !ok || l == nil {
		return nullLogger
	}
	// Add grpc_ctxtags tags metadata until now.
	fields := TagsToFields(ctx)
	// Add logger fields added until now.
	for k, v := range l.fields {
		fields[k] = v
	}
	return l.log.WithFields(fields)
}

// TagsToFields transforms the Tags on the supplied context into logger fields.
func TagsToFields(ctx context.Context) map[string]interface{} {
	return grpc_ctxtags.Extract(ctx).Values()
}

// ToContext adds the log.Logger to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, logger *log.Wrapper) context.Context {
	l := &ctxLogger{
		logger: logger,
	}
	return context.WithValue(ctx, ctxMarkerKey, l)
}

// Debug is equivalent to calling Debug on the log.Logger in the context.
// It is a no-op if the context does not contain a log.Logger.
func Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	Extract(ctx).Fields(fields).Log(log.DebugLevel, msg)
}

// Info is equivalent to calling Info on the log.Logger in the context.
// It is a no-op if the context does not contain a log.Logger.
func Info(ctx context.Context, msg string, fields map[string]interface{}) {
	Extract(ctx).Fields(fields).Log(log.InfoLevel, msg)
}

// Warn is equivalent to calling Warn on the log.Logger in the context.
// It is a no-op if the context does not contain a log.Logger.
func Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	Extract(ctx).Fields(fields).Log(log.WarnLevel, msg)
}

// Error is equivalent to calling Error on the log.Logger in the context.
// It is a no-op if the context does not contain a log.Logger.
func Error(ctx context.Context, msg string, fields map[string]interface{}) {
	Extract(ctx).Fields(fields).Log(log.ErrorLevel, msg)
}
