package xdb

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
)

func getError(err error, query string, args []interface{}) error {
	return fmt.Errorf("%w(sql:%s,args:%+v)", err, query, args)
}

func resolveRows(rows *sql.Rows, col int) (dataRows Rows, err error) {
	dataRows = NewRows()
	columnTypes, _ := rows.ColumnTypes()
	values := make([]interface{}, len(columnTypes))

	for rows.Next() {
		prepareValues(values, columnTypes)
		rows.Scan(values...)
		mapValue := map[string]interface{}{}
		scanIntoMap(mapValue, values, columnTypes)
		dataRows = append(dataRows, mapValue)
	}
	rows.Close()
	return
}

func prepareValues(values []interface{}, columnTypes []*sql.ColumnType) {
	if len(columnTypes) > 0 {
		for idx, columnType := range columnTypes {
			//fmt.Println("idx:", idx, columnType.Name(), columnType.ScanType())
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

func scanIntoMap(mapValue map[string]interface{}, values []interface{}, columnTypes []*sql.ColumnType) {

	for idx, colType := range columnTypes {
		column := colType.Name()
		//fmt.Println("column:", column)
		if reflectValue := reflect.Indirect(reflect.Indirect(reflect.ValueOf(values[idx]))); reflectValue.IsValid() {
			//fmt.Println("scanIntoMap.1")
			mapValue[column] = reflectValue.Interface()
			if valuer, ok := mapValue[column].(driver.Valuer); ok {
				//fmt.Println("scanIntoMap.1.1", valuer)
				mapValue[column], _ = valuer.Value()
			} else if b, ok := mapValue[column].(sql.RawBytes); ok {
				//fmt.Println("scanIntoMap.1.2", string(b))
				mapValue[column] = string(b)
			}
		} else {
			//fmt.Println("scanIntoMap.2")
			mapValue[column] = nil
		}
	}
}
