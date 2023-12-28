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

// 解析数据结果
func ResolveFirstDataResult(rows *sql.Rows, result any) (err error) {
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
		prepareValues(values, columnTypes)

		if rows.Next() {
			err = rows.Scan(values...)
			if err != nil {
				return
			}
			err = scanIntoMap(mapval, values, columns)
		}

	case reflect.Struct:
		fields := cachedTypeFields(reflect.Indirect(rv).Type())

		columnTypes, _ := rows.ColumnTypes()
		columns, _ := rows.Columns()
		values := make([]interface{}, len(columnTypes))
		prepareValues(values, columnTypes)
		if rows.Next() {
			err = rows.Scan(values...)
			if err != nil {
				return
			}
			err = scanInToStruct(fields, rv, columns, values)
		}
	default:
		return &xdb.InvalidArgTypeError{Type: rv.Type()}
	}
	return
}

// 解析数据结果
func ResolveRowsDataResult(rows *sql.Rows, result any) (err error) {

	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Pointer {
		return &xdb.InvalidArgTypeError{Type: rv.Type()}
	}
	if !(rv.Elem().Kind() == reflect.Array ||
		rv.Elem().Kind() == reflect.Slice) {
		return &xdb.InvalidArgTypeError{Type: rv.Elem().Type()}
	}
	rv = rv.Elem()
	var reflectResults reflect.Value
	reflectResults, err = resolveRows(rows, rv)
	if err != nil {
		return
	}
	rv.Set(reflectResults)
	return
}

func ResolveScalar(rows *sql.Rows) (val any, err error) {
	columnTypes, _ := rows.ColumnTypes()
	values := make([]interface{}, len(columnTypes))
	prepareValues(values, columnTypes)
	if rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return
		}
		val = values[0]
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

		err = scanIntoMap(reflect.ValueOf(mapValue), values, columns)
		return mapValue, err
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
		err = scanIntoMap(reflect.ValueOf(mapValue), values, columns)
		if err != nil {
			return
		}
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
func scanInToStruct(fields *structFields, rv reflect.Value, cols []string, vals []any) (err error) {

	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &xdb.InvalidArgTypeError{Type: rv.Type()}
	}

	if rv.CanConvert(structScannerType) {
		scanner := rv.Interface().(xdb.StructScanner)
		err = scanner.StructScan(vals...)
		return
	}

	rv = rv.Elem()

	for i := range cols {
		col := cols[i]
		field, ok := fields.exactName[col]
		if !ok {
			continue
		}
		fv := rv.Field(field.index)
		err = field.dencoder(fv, reflect.Indirect(reflect.ValueOf(vals[i])).Interface())
		if err != nil {
			err = xdb.NewError(fmt.Errorf("field:%s,val:%+v,err:%w", field.name, vals[i], err), "", nil)
			return
		}
	}
	return nil
}

func resolveRows(rows *sql.Rows, rv reflect.Value) (reflectResults reflect.Value, err error) {
	itemType := reflect.Indirect(rv).Type().Elem()

	var kind reflect.Kind = itemType.Kind()

	switch {
	case kind == reflect.Map ||
		(kind == reflect.Ptr && itemType.Elem().Kind() == reflect.Map):
		reflectResults, err = resolveRowsToMap(rows, itemType)
	case kind == reflect.Struct ||
		(kind == reflect.Ptr && itemType.Elem().Kind() == reflect.Struct):
		reflectResults, err = resolveRowsToStruct(rows, itemType)
	default:
		err = &xdb.InvalidArgTypeError{Type: rv.Type()}
		return
	}
	return
}

func resolveRowsToStruct(rows *sql.Rows, itemType reflect.Type) (reflectResults reflect.Value, err error) {
	reflectResults = reflect.MakeSlice(reflect.SliceOf(itemType), 0, 1)

	isPtr := false
	if itemType.Kind() == reflect.Pointer {
		isPtr = true
		itemType = itemType.Elem()
	}

	fields := cachedTypeFields(itemType)
	columnTypes, _ := rows.ColumnTypes()
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columnTypes))
	prepareValues(values, columnTypes)
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

func resolveRowsToMap(rows *sql.Rows, itemType reflect.Type) (reflectResults reflect.Value, err error) {
	reflectResults = reflect.MakeSlice(reflect.SliceOf(itemType), 0, 1)
	isPtr := false
	if itemType.Kind() == reflect.Pointer {
		isPtr = true
		itemType = itemType.Elem()
	}

	columnTypes, _ := rows.ColumnTypes()
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columnTypes))
	prepareValues(values, columnTypes)

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

// func resolveRowsData(rows *sql.Rows, reflectResults reflect.Value, itemType reflect.Type) (err error) {
// 	switch itemType.Kind() {
// 	case reflect.Map:

// 		columnTypes, _ := rows.ColumnTypes()
// 		columns, _ := rows.Columns()
// 		values := make([]interface{}, len(columnTypes))
// 		prepareValues(values, columnTypes)

// 		for rows.Next() {
// 			err = rows.Scan(values...)
// 			if err != nil {
// 				return
// 			}
// 			// 创建一个新的 map 实例，键和值的类型是 result 中 map 的类型
// 			mapval := reflect.MakeMap(itemType)
// 			err = scanIntoMap(mapval, values, columns)
// 			if err != nil {
// 				return
// 			}
// 			reflectResults = reflect.Append(reflectResults, mapval)
// 		}

// 	case reflect.Struct:
// 		fields := cachedTypeFields(itemType)

// 		columnTypes, _ := rows.ColumnTypes()
// 		columns, _ := rows.Columns()
// 		values := make([]interface{}, len(columnTypes))
// 		prepareValues(values, columnTypes)
// 		for rows.Next() {
// 			err = rows.Scan(values...)
// 			if err != nil {
// 				return
// 			}
// 			itemVal := reflect.New(itemType)
// 			err = scanInToStruct(fields, itemVal, columns, values)
// 			if err != nil {
// 				return
// 			}
// 			reflectResults = reflect.Append(reflectResults, itemVal.Elem())
// 		}
// 	case reflect.Pointer:
// 		return resolveRowsData(rows, reflectResults, itemType.Elem())
// 	default:
// 		return &xdb.InvalidArgTypeError{Type: reflectResults.Type()}
// 	}
// 	return nil
// }

// func parseResultSchema(dest any) (schema *Schema, err error) {
// 	if dest == nil {
// 		return nil, fmt.Errorf("目标对象为null: %+v", dest)
// 	}

// 	value := reflect.ValueOf(dest)
// 	if value.Kind() == reflect.Ptr && value.IsNil() {
// 		value = reflect.New(value.Type().Elem())
// 	}
// 	modelType := reflect.Indirect(value).Type()

// 	if modelType.Kind() == reflect.Interface {
// 		modelType = reflect.Indirect(reflect.ValueOf(dest)).Elem().Type()
// 	}

// 	for modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Ptr {
// 		modelType = modelType.Elem()
// 	}

// 	if modelType.Kind() != reflect.Struct {
// 		if modelType.PkgPath() == "" {
// 			return nil, fmt.Errorf("%s: %+v", modelType.Name(), dest)
// 		}
// 		return nil, fmt.Errorf("%s.%s", modelType.PkgPath(), modelType.Name())
// 	}

// 	if v, ok := schemaCache.Load(modelType); ok {
// 		s := v.(*Schema)
// 		// Wait for the initialization of other goroutines to complete
// 		<-s.initialized
// 		return s, s.err
// 	}

// 	schema = &Schema{
// 		Name:         modelType.Name(),
// 		ModelType:    modelType,
// 		FieldsByName: map[string]*field{},
// 		initialized:  make(chan struct{}),
// 	}

// 	// Cache the schema
// 	if v, loaded := schemaCache.LoadOrStore(modelType, schema); loaded {
// 		s := v.(*Schema)
// 		// Wait for the initialization of other goroutines to complete
// 		<-s.initialized
// 		return s, s.err
// 	}

// 	// When the schema initialization is completed, the channel will be closed
// 	defer close(schema.initialized)

// 	for i := 0; i < modelType.NumField(); i++ {
// 		fieldStruct := modelType.Field(i)
// 		if ast.IsExported(fieldStruct.Name) {
// 			continue
// 		}
// 		field := schema.ParseField(fieldStruct)
// 		schema.Fields = append(schema.Fields, field)
// 	}
// 	return
// }
