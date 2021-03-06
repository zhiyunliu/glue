package xdb

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"

	"github.com/zhiyunliu/glue/xdb"
)

func getError(err error, query string, args []interface{}) error {
	return fmt.Errorf("%w(sql:%s,args:%+v)", err, query, args)
}

func resolveRows(rows *sql.Rows) (dataRows xdb.Rows, err error) {
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

func resolveMultiRows(rows *sql.Rows) (datasetRows []xdb.Rows, err error) {
	var setRows xdb.Rows
	for {
		setRows, err = resolveRows(rows)
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
	if len(columnTypes) > 0 {
		for idx, columnType := range columnTypes {
			if columnType.ScanType() != nil {
				values[idx] = reflect.New(reflect.PtrTo(columnType.ScanType())).Interface()
			} else {
				values[idx] = new(interface{})
			}
		}
	} else {
		for idx := range columnTypes {
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
