package pkgs

import "time"

const (
	RequestIDKey  = "x-request-id"
	LoggerKey     = "velocity-logger-request"
	JwtPayloadKey = "jwt_payload"
	IdentityKey   = "identity_user"
)

func GetCurrentTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
