package xdb

type DbError interface {
	Error() string
	Inner() error
	SQL() string
	Args() []interface{}
}

type xDBError struct {
	innerError error
	sql        string
	args       []interface{}
}

func NewError(err error, sql string, args []interface{}) error {
	return &xDBError{
		innerError: err,
		sql:        sql,
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
