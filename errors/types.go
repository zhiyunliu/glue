// nolint:gomnd
package errors

import "net/http"

func BadRequest(reason, message string, opts ...Option) Error {
	return New(http.StatusBadRequest, message)
}

func IsBadRequest(err error) bool {
	return Code(err) == http.StatusBadRequest
}

func Unauthorized(message string, opts ...Option) Error {
	return New(http.StatusUnauthorized, message)
}

func IsUnauthorized(err error) bool {
	return Code(err) == http.StatusUnauthorized
}

func Forbidden(message string, opts ...Option) Error {
	return New(http.StatusForbidden, message)
}

func IsForbidden(err error) bool {
	return Code(err) == http.StatusForbidden
}

func NotFound(message string, opts ...Option) Error {
	return New(http.StatusNotFound, message)
}

func IsNotFound(err error) bool {
	return Code(err) == http.StatusNotFound
}

func InternalServer(message string, opts ...Option) Error {
	return New(http.StatusInternalServerError, message)
}

// IsInternalServer determines if err is an error which indicates an Internal error.
// It supports wrapped errors.
func IsInternalServer(err error) bool {
	return Code(err) == http.StatusInternalServerError
}
