package jwt

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/golibs/xpath"

	"github.com/golang-jwt/jwt/v4"

	gluejwt "github.com/zhiyunliu/glue/auth/jwt"

	"github.com/zhiyunliu/glue/middleware"
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
				token, ok := gluejwt.FromContext(ctx.Context())
				if !ok {
					return reply
				}
				err := writeAuth(ctx, opts, token)
				if err != nil {
					return err
				}
				return reply
			}

			if keyFunc == "" {
				return gluejwt.ErrMissingKeyFunc
			}
			authVal := ctx.Header(gluejwt.AuthorizationKey)

			auths := strings.SplitN(authVal, " ", 2)
			if len(auths) != 2 || !strings.EqualFold(auths[0], gluejwt.BearerWord) {
				return gluejwt.ErrMissingJwtToken
			}
			jwtToken := auths[1]
			tokenData, err := gluejwt.Verify(jwtToken, opts.Secret)
			if err != nil {
				return err
			}
			nctx := gluejwt.NewContext(ctx.Context(), tokenData)
			ctx.ResetContext(nctx)
			return handler(ctx)
		}

	}
}

func writeAuth(ctx context.Context, opts *options, data map[string]interface{}) error {
	tokenVal, err := gluejwt.Sign(opts.signingMethod.Alg(), opts.Secret, data, int64(opts.Expire))
	if err != nil {
		return err
	}
	ctx.Response().Header(gluejwt.AuthorizationKey, fmt.Sprintf("%s %s", gluejwt.BearerWord, tokenVal))
	return nil
}
