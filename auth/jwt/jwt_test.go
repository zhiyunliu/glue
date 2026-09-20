package jwt

import (
	"reflect"
	"testing"

	jwt "github.com/golang-jwt/jwt/v4"
)

func Test_getSecret(t *testing.T) {
	type args struct {
		secret interface{}
		claims jwt.Claims
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Test string secret",
			args: args{
				secret: "mysecret",
				claims: jwt.MapClaims{},
			},
			want:    []byte("mysecret"),
			wantErr: false,
		},
		{
			name: "Test function secret",
			args: args{
				secret: func(c jwt.Claims) ([]byte, error) {
					return []byte("functionsecret"), nil
				},
				claims: jwt.MapClaims{},
			},
			want:    []byte("functionsecret"),
			wantErr: false,
		},
		{
			name: "Test Keyfunc secret",
			args: args{
				secret: Keyfunc(func(c jwt.Claims) ([]byte, error) {
					return []byte("keyfuncsecret"), nil
				}),
				claims: jwt.MapClaims{},
			},
			want:    []byte("keyfuncsecret"),
			wantErr: false,
		},
		{
			name: "Test unknown secret",
			args: args{
				secret: 123,
				claims: jwt.MapClaims{},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSecret(tt.args.secret, tt.args.claims)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerify(t *testing.T) {

	type args struct {
		tokenVal string
		secret   interface{}
	}

	validToken, _ := Sign("HS256", "secret_key", map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}, 60)

	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "ValidToken",
			args: args{
				tokenVal: validToken,
				secret:   "secret_key",
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},

			wantErr: false,
		},
		{
			name: "InvalidToken",
			args: args{
				tokenVal: "invalid_token",
				secret:   "secret_key",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ExpiredToken",
			args: args{
				tokenVal: "expired_token",
				secret:   "secret_key",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "TokenWithUnsupportedSigningMethod",
			args: args{
				tokenVal: "unsupported_token",
				secret:   "secret_key",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Verify(tt.args.tokenVal, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestSign(t *testing.T) {
	type args struct {
		signingMethod string
		secret        interface{}
		data          map[string]interface{}
		timeout       int64
		opts          []Option
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Valid case with timeout",
			args: args{
				signingMethod: "HS256",
				secret:        "my-secret",
				data: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
				timeout: 60,
			},
			wantErr: false,
		},
		{
			name: "Valid case without timeout",
			args: args{
				signingMethod: "HS256",
				secret:        "my-secret",
				data: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
				timeout: 0,
			},
			wantErr: false,
		},
		{
			name: "Invalid signing method",
			args: args{
				signingMethod: "InvalidMethod",
				secret:        "my-secret",
				data: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
				timeout: 60,
			},
			wantErr: true,
		},
		{
			name: "Invalid secret",
			args: args{
				signingMethod: "HS256",
				secret:        12345,
				data: map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
				timeout: 60,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Sign(tt.args.signingMethod, tt.args.secret, tt.args.data, tt.args.timeout, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

}
