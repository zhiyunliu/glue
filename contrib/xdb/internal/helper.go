package internal

import (
	"database/sql"
	"database/sql/driver"
	"reflect"

	"github.com/zhiyunliu/glue/xdb"
)

func Unwrap(args ...interface{}) []interface{} {
	nargs := make([]interface{}, len(args))
	for i := range args {
		rv := reflect.ValueOf(args[i])
		if rv.Kind() == reflect.Ptr {
			nargs[i] = rv.Elem().Interface()
			continue
		}
		nargs[i] = args[i]
	}
	return nargs
}

func GetError(err error, query string, args ...interface{}) error {
	return xdb.NewError(err, query, Unwrap(args...))
}

func ResolveRows(rows *sql.Rows) (dataRows xdb.Rows, err error) {
	dataRows = xdb.NewRows()
	columnTypes, _ := rows.ColumnTypes()
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columnTypes))
	prepareValues(values, columnTypes)
	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return
		}
		mapValue := map[string]interface{}{}
		scanIntoMap(mapValue, values, columns)
		dataRows = append(dataRows, mapValue)
	}
	return
}

func ResolveMultiRows(rows *sql.Rows) (datasetRows []xdb.Rows, err error) {
	var setRows xdb.Rows
	for {
		setRows, err = ResolveRows(rows)
		if err != nil {
			return
		}
		datasetRows = append(datasetRows, setRows)
		if !rows.NextResultSet() {
			return
		}
	}
}

func prepareValues(values []interface{}, columnTypes []*sql.ColumnType) {
	for idx, columnType := range columnTypes {
		t, ok := getDbType(columnType.DatabaseTypeName())
		if !ok {
			t = columnType.ScanType()
		}
		if t != nil {
			values[idx] = reflect.New(reflect.PtrTo(t)).Interface()
		} else {
			values[idx] = new(interface{})
		}
	}
}

func scanIntoMap(mapValue map[string]interface{}, values []interface{}, columnTypes []string) {

	for idx, column := range columnTypes {
		if reflectValue := reflect.Indirect(reflect.Indirect(reflect.ValueOf(values[idx]))); reflectValue.IsValid() {
			mapValue[column] = reflectValue.Interface()
			if valuer, ok := mapValue[column].(driver.Valuer); ok {
				mapValue[column], _ = valuer.Value()
			} else if b, ok := mapValue[column].(sql.RawBytes); ok {
				mapValue[column] = string(b)
			}
		} else {
			mapValue[column] = nil
		}
	}
}
