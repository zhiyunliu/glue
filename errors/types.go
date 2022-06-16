// nolint:gomnd
package errors

func BadRequest(reason, message string) *Error {
	return New(400, message)
}

func IsBadRequest(err error) bool {
	return Code(err) == 400
}

func Unauthorized(message string) *Error {
	return New(401, message)
}

func IsUnauthorized(err error) bool {
	return Code(err) == 401
}

func Forbidden(message string) *Error {
	return New(403, message)
}

func IsForbidden(err error) bool {
	return Code(err) == 403
}

func NotFound(message string) *Error {
	return New(404, message)
}

func IsNotFound(err error) bool {
	return Code(err) == 404
}

func InternalServer(message string) *Error {
	return New(500, message)
}
