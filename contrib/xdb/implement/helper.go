package implement

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xreflect"
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

// 解析数据结果
func ResolveFirstDataResult(proto string, rows *sql.Rows, result any) (err error) {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &xdb.InvalidArgTypeError{Type: rv.Type()}
	}

	indirectType := reflect.Indirect(rv).Type()
	switch indirectType.Kind() {
	case reflect.Map:
		// 创建一个新的 map 实例，键和值的类型是 result 中 map 的类型
		mapval := reflect.MakeMapWithSize(indirectType, 0)
		// 将新 map 设置回 result 指向的位置
		rv.Elem().Set(mapval)

		columnTypes, _ := rows.ColumnTypes()
		columns, _ := rows.Columns()
		values := make([]interface{}, len(columnTypes))
		prepareValues(proto, values, columnTypes)

		if rows.Next() {
			err = rows.Scan(values...)
			if err != nil {
				return
			}
			err = scanIntoMap(mapval, values, columns)
		} else {
			return xdb.EmptyError
		}

	case reflect.Struct:
		fields := xreflect.CachedTypeFields(reflect.Indirect(rv).Type())

		columnTypes, _ := rows.ColumnTypes()
		columns, _ := rows.Columns()
		values := make([]interface{}, len(columnTypes))
		prepareValues(proto, values, columnTypes)
		if rows.Next() {
			err = rows.Scan(values...)
			if err != nil {
				return
			}
			err = scanInToStruct(fields, rv, columns, values)
		} else {
			return xdb.EmptyError
		}
	default:
		return &xdb.InvalidArgTypeError{Type: rv.Type()}
	}
	return
}

// 解析数据结果
func ResolveRowsDataResult(proto string, rows *sql.Rows, result any) (err error) {

	rv := reflect.ValueOf(result)

	if reader := xdb.GetRowDataReader(result); reader != nil {
		return resolveRowsToReader(proto, rows, reader)
	}

	if rv.Kind() != reflect.Pointer {
		return &xdb.InvalidArgTypeError{Type: rv.Type()}
	}
	if !(rv.Elem().Kind() == reflect.Array ||
		rv.Elem().Kind() == reflect.Slice) {
		return &xdb.InvalidArgTypeError{Type: rv.Elem().Type()}
	}
	rv = rv.Elem()
	var reflectResults reflect.Value
	reflectResults, err = resolveRows(proto, rows, rv)
	if err != nil {
		return
	}
	rv.Set(reflectResults)
	return
}

func ResolveScalar(proto string, rows *sql.Rows) (val any, err error) {
	columnTypes, _ := rows.ColumnTypes()
	values := make([]interface{}, len(columnTypes))
	prepareValues(proto, values, columnTypes)
	if rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return
		}
		val = values[0]
	}
	if val == nil {
		return
	}
	rv := reflect.ValueOf(val)
	for !rv.IsZero() && rv.Type().Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.CanInterface() {
		return rv.Interface(), nil
	}
	return nil, nil
}

func ResolveFirstRow(proto string, rows *sql.Rows) (dataRows xdb.Row, err error) {
	dataRows = xdb.NewRow()
	columnTypes, _ := rows.ColumnTypes()
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columnTypes))
	prepareValues(proto, values, columnTypes)
	if rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return
		}
		mapValue := map[string]interface{}{}

		err = scanIntoMap(reflect.ValueOf(mapValue), values, columns)
		return mapValue, err
	}
	return
}

func ResolveRows(proto string, rows *sql.Rows) (dataRows xdb.Rows, err error) {
	dataRows = xdb.NewRows()
	columnTypes, _ := rows.ColumnTypes()
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columnTypes))
	prepareValues(proto, values, columnTypes)
	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return
		}
		mapValue := map[string]interface{}{}
		err = scanIntoMap(reflect.ValueOf(mapValue), values, columns)
		if err != nil {
			return
		}
		dataRows = append(dataRows, mapValue)
	}
	return
}

func ResolveMultiRows(proto string, rows *sql.Rows) (datasetRows []xdb.Rows, err error) {
	var setRows xdb.Rows
	for {
		setRows, err = ResolveRows(proto, rows)
		if err != nil {
			return
		}
		datasetRows = append(datasetRows, setRows)
		if !rows.NextResultSet() {
			return
		}
	}
}

func ResolveParams(input any, callback xdb.StmtDbTypeWrap) (params xtypes.XMap, err error) {
	if input == nil {
		return xtypes.XMap{}, nil
	}
	rv := reflect.ValueOf(input)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return xtypes.XMap{}, nil
	}

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
	case xdb.DbParamConverter:
		return t.ToDbParam(), nil
	default:
		return analyzeParamFields(t, callback)
	}

}

func analyzeParamFields(input any, callback xdb.StmtDbTypeWrap) (params xtypes.XMap, err error) {
	params = make(xtypes.XMap)
	refval := reflect.ValueOf(input)
	//获取最终的类型值
	for refval.Kind() == reflect.Pointer {
		refval = refval.Elem()
	}

	if refval.Kind() != reflect.Struct {
		return params, fmt.Errorf("只能接收struct; 实际是 %s", refval.Kind().String())
	}

	fields := xreflect.CachedTypeFields(refval.Type())

	for _, f := range fields.ExactName {
		if val, ok := f.Encoder(refval); ok {
			if callback != nil {
				val = callback(val, f.TagOpts)
			}
			params[f.Name] = val
		}
	}
	return
}

func prepareValues(proto string, values []interface{}, columnTypes []*sql.ColumnType) {
	for idx, columnType := range columnTypes {
		t, ok := xdb.GetDbType(proto, columnType.DatabaseTypeName())
		if !ok {
			t = columnType.ScanType()
		}
		if t != nil {
			values[idx] = reflect.New(reflect.PointerTo(t)).Interface()
		} else {
			values[idx] = new(interface{})
		}
	}
}

func scanIntoMap(mapValue reflect.Value, values []interface{}, columnTypes []string) (err error) {
	var val any
	for idx, column := range columnTypes {
		if reflectValue := reflect.Indirect(reflect.ValueOf(values[idx])); reflectValue.IsValid() {
			val = reflectValue.Interface()
			if valuer, ok := val.(driver.Valuer); ok {
				if reflect.ValueOf(valuer).IsNil() {
					val = nil
				} else {
					val, err = valuer.Value()
				}
				if err != nil {
					return
				}
			} else if b, ok := val.(sql.RawBytes); ok {
				val = string(b)
			}
		} else {
			val = nil
		}
		mapValue.SetMapIndex(reflect.ValueOf(column), reflect.Indirect(reflect.ValueOf(val)))
	}
	return nil
}

// 填充数据到结构体
func scanInToStruct(fields *xreflect.StructFields, rv reflect.Value, cols []string, vals []any) (err error) {

	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &xdb.InvalidArgTypeError{Type: rv.Type()}
	}

	//获取最终的struct 类型
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	for i := range cols {
		col := cols[i]
		//	field, ok := fields.ExactName[col]

		vrf := reflect.ValueOf(vals[i])
		for vrf.Kind() == reflect.Ptr {
			vrf = vrf.Elem()
		}

		if !(vrf.IsValid() && vrf.CanInterface()) {
			continue
		}
		err = fields.Dencode(rv, col, vrf.Interface())
		if err != nil {
			err = xdb.NewError(fmt.Errorf("field:%s,val:%+v,err:%w", col, vals[i], err), "", nil)
			return
		}

	}
	return nil
}

func resolveRows(proto string, rows *sql.Rows, rv reflect.Value) (reflectResults reflect.Value, err error) {
	itemType := reflect.Indirect(rv).Type().Elem()

	var kind reflect.Kind = itemType.Kind()

	switch {
	case kind == reflect.Map ||
		(kind == reflect.Ptr && itemType.Elem().Kind() == reflect.Map):
		reflectResults, err = resolveRowsToMap(proto, rows, itemType)
	case kind == reflect.Struct ||
		(kind == reflect.Ptr && itemType.Elem().Kind() == reflect.Struct):
		reflectResults, err = resolveRowsToStruct(proto, rows, itemType)
	default:
		err = &xdb.InvalidArgTypeError{Type: rv.Type()}
		return
	}
	return
}

func resolveRowsToStruct(proto string, rows *sql.Rows, itemType reflect.Type) (reflectResults reflect.Value, err error) {
	reflectResults = reflect.MakeSlice(reflect.SliceOf(itemType), 0, 1)

	isPtr := false
	if itemType.Kind() == reflect.Pointer {
		isPtr = true
		itemType = itemType.Elem()
	}

	fields := xreflect.CachedTypeFields(itemType)
	columnTypes, _ := rows.ColumnTypes()
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columnTypes))
	prepareValues(proto, values, columnTypes)
	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return
		}

		itemVal := reflect.New(itemType)
		err = scanInToStruct(fields, itemVal, columns, values)
		if err != nil {
			return
		}
		if isPtr {
			reflectResults = reflect.Append(reflectResults, itemVal)
		} else {
			reflectResults = reflect.Append(reflectResults, itemVal.Elem())
		}
	}
	return
}

func resolveRowsToMap(proto string, rows *sql.Rows, itemType reflect.Type) (reflectResults reflect.Value, err error) {
	reflectResults = reflect.MakeSlice(reflect.SliceOf(itemType), 0, 1)
	isPtr := false
	if itemType.Kind() == reflect.Pointer {
		isPtr = true
		itemType = itemType.Elem()
	}

	columnTypes, _ := rows.ColumnTypes()
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columnTypes))
	prepareValues(proto, values, columnTypes)

	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return
		}
		// 创建一个新的 map 实例，键和值的类型是 result 中 map 的类型
		mapval := reflect.MakeMap(itemType)
		err = scanIntoMap(mapval, values, columns)
		if isPtr {
			mapPtr := reflect.New(itemType)
			mapPtr.Elem().Set(mapval)
			reflectResults = reflect.Append(reflectResults, mapPtr)
		} else {
			reflectResults = reflect.Append(reflectResults, mapval)
		}
	}
	return
}

func resolveRowsToReader(proto string, rows *sql.Rows, reader xdb.RowDataReader) (err error) {
	rowItem := reader.GetRowItem()
	defer reader.Close()
	for {
		err = ResolveFirstDataResult(proto, rows, rowItem)
		if err != nil {
			if errors.Is(err, xdb.EmptyError) {
				return nil
			}
			return
		}
		if err = reader.FillRowItem(rowItem); err != nil {
			return
		}
	}
}
