package implement

import "database/sql"

type DbResolveResultCallback func(*sql.Rows, any) error
type DbResolveMapValCallback func(*sql.Rows) (result any, err error)
