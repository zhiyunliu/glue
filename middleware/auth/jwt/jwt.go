package jwt

import (
	"fmt"
	"strings"
	"time"

	sysctx "context"

	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/golibs/xpath"

	"github.com/golang-jwt/jwt/v4"

	"github.com/zhiyunliu/gel/errors"
	"github.com/zhiyunliu/gel/middleware"
)

type authKey struct{}

const (

	// bearerWord the bearer key word for authorization
	bearerWord string = "Bearer"

	// authorizationKey holds the key used to store the JWT Token in the request tokenHeader.
	authorizationKey string = "Authorization"
)

var (
	ErrMissingJwtToken        = errors.Unauthorized("JWT token is missing")
	ErrMissingKeyFunc         = errors.Unauthorized("secret is missing")
	ErrTokenInvalid           = errors.Unauthorized("Token is invalid")
	ErrTokenExpired           = errors.Unauthorized("JWT token has expired")
	ErrTokenParseFail         = errors.Unauthorized("Fail to parse JWT token ")
	ErrUnSupportSigningMethod = errors.Unauthorized("Wrong signing method")
	ErrWrongContext           = errors.Unauthorized("Wrong context for middleware")
	ErrNeedTokenProvider      = errors.Unauthorized("Token provider is missing")
	ErrSignToken              = errors.Unauthorized("Can not sign token.Is the key correct?")
	ErrGetKey                 = errors.Unauthorized("Can not get key while signing token")
)

// Option is jwt option.
type Option func(*options)

// Parser is a jwt parser
type options struct {
	signingMethod jwt.SigningMethod
	Secret        string
	Excludes      []string
	Expire        int //单位：second
}

// WithSigningMethod with signing method option.
func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

func WithSecret(secret string) Option {
	return func(o *options) {
		o.Secret = secret
	}
}

func WithExcludes(excludes ...string) Option {
	return func(o *options) {
		o.Excludes = excludes
	}
}

func Server(opts ...Option) middleware.Middleware {
	o := &options{
		signingMethod: jwt.SigningMethodHS256,
		Expire:        60 * 60,
	}
	for _, opt := range opts {
		opt(o)
	}
	return serverByOptions(o)
}

func serverByConfig(cfg *Config) middleware.Middleware {
	opts := &options{
		signingMethod: jwt.GetSigningMethod(cfg.Method),
		Secret:        cfg.Secret,
		Excludes:      cfg.Excludes,
		Expire:        cfg.Expire,
	}
	return serverByOptions(opts)
}

func serverByOptions(opts *options) middleware.Middleware {
	keyFunc := opts.Secret
	excludeMatch := xpath.NewMatch(opts.Excludes...)

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			path := ctx.Request().Path().GetURL().Path

			isMatch, _ := excludeMatch.Match(path, "/")
			if isMatch {
				//是排除路径，不进行jwt检查
				reply = handler(ctx)
				if tmpdata, ok := ctx.Meta()["jwt_data"]; ok {
					if data, ok := tmpdata.(map[string]interface{}); ok {
						write(ctx, opts, data)
					}
				}
				return reply
			}

			if keyFunc == "" {
				return ErrMissingKeyFunc
			}
			authVal := ctx.Header(authorizationKey)

			auths := strings.SplitN(authVal, " ", 2)
			if len(auths) != 2 || !strings.EqualFold(auths[0], bearerWord) {
				return ErrMissingJwtToken
			}
			jwtToken := auths[1]
			tokenData, err := Verify(jwtToken, opts.Secret)
			if err != nil {
				return err
			}
			ctx = NewContext(ctx, tokenData)
			return handler(ctx)
		}

	}
}

// NewContext put auth info into context
func NewContext(ctx context.Context, info map[string]interface{}) context.Context {
	nctx := sysctx.WithValue(ctx.Context(), authKey{}, info)
	ctx.ResetContext(nctx)
	return ctx
}

// FromContext extract auth info from context
func FromContext(ctx context.Context) (token map[string]interface{}, ok bool) {
	token, ok = ctx.Context().Value(authKey{}).(map[string]interface{})
	return
}

func Verify(tokenVal, secret string) (map[string]interface{}, error) {
	tokenInfo, err := jwt.Parse(tokenVal, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrTokenInvalid
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return nil, ErrTokenExpired
			} else {
				return nil, ErrTokenParseFail
			}
		}
		return nil, errors.Unauthorized(err.Error())
	}
	if !tokenInfo.Valid {
		return nil, ErrTokenInvalid
	}

	if claims, ok := tokenInfo.Claims.(jwt.MapClaims); ok {
		return claims["data"].(map[string]interface{}), nil
	}

	return nil, ErrUnSupportSigningMethod
}

func Sign(signingMethod string, secret string, data map[string]interface{}, timeout int64) (string, error) {
	expireAt := time.Now().Unix() + timeout
	if timeout == 0 {
		expireAt = 0
	}
	claims := &jwt.MapClaims{
		"exp":  expireAt,
		"data": data,
	}
	method := jwt.GetSigningMethod(signingMethod)
	token := jwt.NewWithClaims(method, claims)
	return token.SignedString([]byte(secret))
}

func write(ctx context.Context, opts *options, data map[string]interface{}) error {
	tokenVal, err := Sign(opts.signingMethod.Alg(), opts.Secret, data, int64(opts.Expire))
	if err != nil {
		return err
	}
	ctx.Response().Header(authorizationKey, fmt.Sprintf("%s %s", bearerWord, tokenVal))
	return nil
}

func Write(ctx context.Context, data map[string]interface{}) {
	ctx.Meta()["jwt_data"] = data
}
