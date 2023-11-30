package internal

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xtypes"
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

func ResolveFirstDataResult(rows *sql.Rows, result any) (err error) {
	for rows.Next() {
		err = rows.Scan(result)
		return
	}
	return
}

// 解析单个数据结果
func ResolveDataResult(rows *sql.Rows, result any) (err error) {
	for rows.Next() {
		err = rows.Scan(result)
		return
	}
	return
}

// 解析多个数据结果
func ResolveMultiDataResult(rows *sql.Rows, result []any) (err error) {
	for rows.Next() {
		err = rows.Scan(result)
		return
	}
	return
}

func ResolveScalar(rows *sql.Rows) (val any, err error) {
	val = new(interface{})
	if rows.Next() {
		err = rows.Scan(val)
		if err != nil {
			return
		}
		return val, nil
	}
	return
}

func ResolveFirstRow(rows *sql.Rows) (dataRows xdb.Row, err error) {
	columnTypes, _ := rows.ColumnTypes()
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columnTypes))
	prepareValues(values, columnTypes)
	if rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return
		}
		mapValue := map[string]interface{}{}
		scanIntoMap(mapValue, values, columns)
		return mapValue, nil
	}
	return
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

func ResolveParams(input any) (params xtypes.XMap, err error) {

	switch t := input.(type) {
	case map[string]any:
		return t, nil
	case xtypes.XMap:
		return t, nil
	case map[string]string:
		params = make(xtypes.XMap)
		for k, v := range t {
			params[k] = v
		}
		return params, nil
	case xtypes.SMap:
		params = make(xtypes.XMap)
		for k, v := range t {
			params[k] = v
		}
		return params, nil
	case xdb.DbParam:
		return t.ToDbParam(), nil
	default:
		return analyzeParamFields(t)
	}

}

func analyzeParamFields(input any) (params xtypes.XMap, err error) {
	params = make(xtypes.XMap)
	refval := reflect.ValueOf(input)
	//获取最终的类型值
	for refval.Kind() == reflect.Pointer {
		refval = refval.Elem()
	}

	if refval.Kind() != reflect.Struct {
		return params, fmt.Errorf("只能接收struct; 实际是 %s", refval.Kind().String())
	}

	fields := cachedTypeFields(refval.Type())

	for i := range fields.list {
		f := &fields.list[i]
		fv := refval.Field(f.index)
		params[f.name] = f.encoder(fv)
	}
	return
}

func prepareValues(values []interface{}, columnTypes []*sql.ColumnType) {
	for idx, columnType := range columnTypes {
		t, ok := xdb.GetDbType(columnType.DatabaseTypeName())
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
