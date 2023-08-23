package internal

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/zhiyunliu/glue/xdb"
)

type WrapArgs struct {
	Name  string
	Value any
	Out   bool
}

func (a WrapArgs) MarshalJSON() (result []byte, err error) {
	tmp := map[string]any{
		a.Name: a.Value,
	}
	if a.Out {
		tmp["out"] = true
	}
	return json.Marshal(tmp)
}

func (a WrapArgs) String() string {
	outv := ""
	if a.Out {
		outv = ",out:true"
	}
	return fmt.Sprintf("{%s:%+v%s}", a.Name, a.Value, outv)
}

func Unwrap(args ...interface{}) []interface{} {
	nargs := make([]interface{}, len(args))
	for i := range args {
		val := args[i]
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Ptr {
			val = rv.Elem().Interface()
		}
		if arg, ok := val.(sql.NamedArg); ok {
			val = arg.Value
			out := false
			if otv, ok := val.(sql.Out); ok {
				out = true
				val = otv.Dest
				orv := reflect.ValueOf(val)
				if orv.Kind() == reflect.Ptr {
					val = orv.Elem().Interface()
				}
			}
			nargs[i] = WrapArgs{
				Name:  arg.Name,
				Value: val,
				Out:   out,
			}
		} else {
			nargs[i] = val
		}
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
