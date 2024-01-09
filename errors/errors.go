package errors

import (
	"errors"
	"fmt"
)

const (
	UnknownCode = 500
)

type IgnoreError interface {
	Ignore() bool
}

type RespError interface {
	GetCode() int
	GetMessage() string
}

type Error struct {
	Code     int               `json:"code,omitempty"`
	Message  string            `json:"message,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (x Error) GetCode() int {
	return x.Code
}

func (x Error) GetMessage() string {
	return x.Message
}

func (x Error) GetMetadata() map[string]string {
	return x.Metadata

}

func (e Error) Error() string {
	return fmt.Sprintf("error: code = %d message = %s metadata = %+v", e.Code, e.Message, e.Metadata)
}

func (e Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code
	}
	return false
}

func (e *Error) WithMetadata(md map[string]string) *Error {
	e.Metadata = md
	return e
}

func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Errorf(code int, format string, a ...interface{}) error {
	return New(code, fmt.Sprintf(format, a...))
}

// Code returns the http code for an error.
// It supports wrapped errors.
func Code(err error) int {
	if err == nil {
		return 200 //nolint:gomnd
	}
	if xerr, ok := err.(*Error); ok {
		return xerr.Code
	}

	return -1
}

// IsInternalServer determines if err is an error which indicates an Internal error.
// It supports wrapped errors.
func IsInternalServer(err error) bool {
	return Code(err) == 500
}

// FromError try to convert an error to *Error.
// It supports wrapped errors.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if se := new(Error); errors.As(err, &se) {
		return se
	}
	return New(UnknownCode, err.Error())
}
