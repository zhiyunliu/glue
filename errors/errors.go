package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var _ Error = (*xError)(nil)

const (
	UnknownCode = 500
)

type IgnoreError interface {
	Ignore() bool
}

type Response interface {
	GetCode() int
	GetSubCode() string
}

type Error interface {
	error
	Response
	GetMessage() string
}

type xError struct {
	Code       int            `json:"code"`
	SubCode    string         `json:"sub_code,omitempty"`
	Message    string         `json:"message"`
	innerErr   error          `json:"-"`
	ErrData    map[string]any `json:"errdata,omitempty"`
	Data       any            `json:"data,omitempty"`
	statusCode int            `json:"-"`
}

func (x xError) GetCode() int {
	return x.Code
}

func (x xError) GetSubCode() string {
	return x.SubCode
}

func (x xError) GetMessage() string {
	return x.Message
}

func (x xError) GetInner() error {
	return x.innerErr
}

func (x xError) GetStatusCode() int {
	return x.statusCode
}

func (x xError) GetData() any {
	return x.Data
}
func (x xError) GetErrData() map[string]any {
	return x.ErrData
}

func (x xError) Error() string {
	return fmt.Sprintf("error:code=%d,subcode=%s,message=%s,errdata=%+v,inner=%s", x.Code, x.SubCode, x.Message, x.ErrData, x.innerErr)
}

func (x xError) Is(err error) bool {
	var tmp Error
	if As(err, &tmp) {
		return tmp.GetCode() == x.Code
	}
	return false
}

// 根据orgErr克隆一个新的Error，并且可以通过opts覆盖原有的字段值。
func Clone(orgErr Error, opts ...Option) Error {
	e := &xError{
		Code:    orgErr.GetCode(),
		Message: orgErr.GetMessage(),
		SubCode: orgErr.GetSubCode(),
	}

	if xerr, ok := orgErr.(*xError); ok {
		e.ErrData = xerr.ErrData
		e.Data = xerr.Data
		e.innerErr = xerr.innerErr
		e.statusCode = xerr.statusCode
	}

	for _, opt := range opts {
		opt(e)
	}
	return e
}

// 创建一个新的Error
func New(code int, message string, opts ...Option) Error {
	e := &xError{
		Code:    code,
		Message: message,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func Errorf(code int, format string, a ...interface{}) error {
	return New(code, fmt.Sprintf(format, a...))
}

// Code returns the http code for an error.
// It supports wrapped errors.
func Code(err error) int {
	if err == nil {
		return http.StatusOK //nolint:gomnd
	}
	if xerr, ok := err.(Error); ok {
		return xerr.GetCode()
	}

	var se Error
	if errors.As(err, &se) {
		return se.GetCode()
	}

	return UnknownCode
}

// FromError try to convert an error to *xError.
// It supports wrapped errors.
func FromError(err error) Error {
	if err == nil {
		return nil
	}
	if xerr, ok := err.(Error); ok {
		return xerr
	}
	var se Error
	if errors.As(err, &se) {
		return se
	}
	return New(UnknownCode, err.Error())
}
