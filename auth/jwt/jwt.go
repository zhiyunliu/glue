package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zhiyunliu/glue/errors"
)

type authKey struct{}

const (

	// bearerWord the bearer key word for authorization
	BearerWord string = "Bearer"

	// authorizationKey holds the key used to store the JWT Token in the request tokenHeader.
	AuthorizationKey string = "Authorization"
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

//timeout 秒
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

// NewContext put auth info into context
func NewContext(ctx context.Context, info map[string]interface{}) context.Context {
	return context.WithValue(ctx, authKey{}, info)
}

// FromContext extract auth info from context
func FromContext(ctx context.Context) (token map[string]interface{}, ok bool) {
	token, ok = ctx.Value(authKey{}).(map[string]interface{})
	return
}
