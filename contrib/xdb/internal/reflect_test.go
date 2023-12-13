package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type MyStruct struct {
	BoolField bool
	IntField1 int
	IntField2 int
	IntField3 int
}

func Test_ReflectSet(t *testing.T) {

	// 创建结构体实例
	var myInstance *MyStruct

	// 创建字段名称和值的映射
	fieldValues := map[string]interface{}{
		"BoolField": true,
		"IntField1": 42,
		"IntField2": 123,
		"IntField3": 987,
	}

	json.Unmarshal([]byte(`{}`), &fieldValues)

	// 使用反射循环为结构体字段赋值
	for fieldName, value := range fieldValues {
		setStructField(myInstance, fieldName, value)
	}

	// 输出赋值后的结构体实例
	fmt.Println(myInstance)
}

// setStructField 使用反射为结构体字段赋值
func setStructField(myStructPtr interface{}, fieldName string, value interface{}) {
	refval := reflect.ValueOf(myStructPtr)
	if refval.IsNil() {
		refval = reflect.New(refval.Type())
	}

	structValue := reflect.Indirect(refval)

	// 获取字段的 reflect.Value
	fieldValue := structValue.FieldByName(fieldName)

	// 如果字段存在且可设置，则设置字段的值
	if fieldValue.IsValid() && fieldValue.CanSet() {
		// 将传入的值转换为 reflect.Value
		newValue := reflect.ValueOf(value)

		// 确保值的类型与字段类型匹配
		if newValue.Type().AssignableTo(fieldValue.Type()) {
			// 设置字段的值
			fieldValue.Set(newValue)
		} else {
			fmt.Printf("Type mismatch for field %s\n", fieldName)
		}
	} else {
		fmt.Printf("Field %s not found or not settable\n", fieldName)
	}
}
