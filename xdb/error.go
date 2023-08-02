package xdb

type DbError interface {
	Error() string
	Inner() error
	SQL() string
	Args() []interface{}
}

type PanicError interface {
	Error() string
	StackTrace() string
}

type xDBError struct {
	innerError error
	sql        string
	stackTrace string
	args       []interface{}
}

func NewError(err error, execSql string, args []interface{}) error {
	return &xDBError{
		innerError: err,
		sql:        execSql,
		args:       args,
	}
}

func (e *xDBError) Error() string {
	return e.innerError.Error()
}

func (e *xDBError) Inner() error {
	return e.innerError
}

func (e *xDBError) SQL() string {
	return e.sql
}

func (e *xDBError) Args() []interface{} {
	return e.args
}

func (e *xDBError) StackTrace() string {
	return e.stackTrace
}

func NewPanicError(err error, strace string) error {
	return &xDBError{
		innerError: err,
		stackTrace: strace,
	}
}
