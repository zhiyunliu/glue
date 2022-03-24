package errors

import (
	"errors"
	"fmt"
)

const (
	UnknownCode = 500
)

type Error struct {
	Code     int               `json:"code,omitempty"`
	Message  string            `json:"message,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (x *Error) GetCode() int {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *Error) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Error) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

//go:generate protoc -I. --go_out=paths=source_relative:. errors.proto

func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d message = %s metadata = %+v", e.Code, e.Message, e.Metadata)
}

// Is matches each error in the chain with the target value.
func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code
	}
	return false
}

// WithMetadata with an MD formed by the mapping of key, value.
func (e *Error) WithMetadata(md map[string]string) *Error {
	e.Metadata = md
	return e
}

// New returns an error object for the code, message.
func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Errorf returns an error object for the code, message and error info.
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
