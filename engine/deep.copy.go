package engine

import "reflect"

func deepCopy[T any](original T) T {
	value := deepCopyValue(reflect.ValueOf(original))
	if !value.IsValid() {
		var zero T
		return zero
	}
	return value.Interface().(T)
}

func deepCopyValue(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return value
	}

	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.New(value.Type().Elem())
		copy.Elem().Set(deepCopyValue(value.Elem()))
		return copy
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := deepCopyValue(value.Elem())
		result := reflect.New(value.Type()).Elem()
		result.Set(copy)
		return result
	case reflect.Struct:
		copy := reflect.New(value.Type()).Elem()
		copy.Set(value)
		for i := 0; i < value.NumField(); i++ {
			if !copy.Field(i).CanSet() || !value.Field(i).CanInterface() {
				continue
			}
			copy.Field(i).Set(deepCopyValue(value.Field(i)))
		}
		return copy
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			copy.Index(i).Set(deepCopyValue(value.Index(i)))
		}
		return copy
	case reflect.Array:
		copy := reflect.New(value.Type()).Elem()
		for i := 0; i < value.Len(); i++ {
			copy.Index(i).Set(deepCopyValue(value.Index(i)))
		}
		return copy
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.MakeMapWithSize(value.Type(), value.Len())
		iter := value.MapRange()
		for iter.Next() {
			copy.SetMapIndex(deepCopyValue(iter.Key()), deepCopyValue(iter.Value()))
		}
		return copy
	default:
		return value
	}
}
