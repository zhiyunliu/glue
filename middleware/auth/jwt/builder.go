package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/zhiyunliu/gel/encoding"
	"github.com/zhiyunliu/gel/middleware"
)

func NewBuilder() middleware.MiddlewareBuilder {
	return &xBuilder{}
}

type xBuilder struct {
}

func (xBuilder) Name() string {
	return "jwt"
}
func (xBuilder) Build(data middleware.RawMessage) middleware.Middleware {

	cfg := &Config{}
	encoding.GetCodec(data.Codec).Unmarshal(data.Data, &cfg)

	return Server(func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
}
