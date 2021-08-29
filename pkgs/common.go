package pkgs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const (
	RequestIDKey  = "x-request-id"
	LoggerKey     = "velocity-logger-request"
	JwtPayloadKey = "jwt_payload"
	IdentityKey   = "identity_user"
)

func GetCurrentTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetRequestID request id from header
func GetRequestID(ctx context.Context) string {
	id := GetHeaderFirst(ctx, RequestIDKey)
	if id == "" {
		id = NewRequestID()
	}
	return id
}

// GetUsername get username from header
func GetUsername(ctx context.Context) string {
	return GetHeaderFirst(ctx, UsernameKey)
}

// GetHeaderFirst get header first value
func GetHeaderFirst(ctx context.Context, key string) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get(key); len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// NewRequestID generate a RequestId
func NewRequestID() string {
	return uuid.New().String()
}
