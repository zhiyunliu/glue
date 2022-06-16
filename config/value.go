package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	_ Value = (*atomicValue)(nil)
	_ Value = (*emptyValue)(nil)
)

// Value is config value interface.
type Value interface {
	Bool() (bool, error)
	Int() (int64, error)
	Float() (float64, error)
	String() string
	Duration() (time.Duration, error)
	Slice() ([]Value, error)
	Map() (map[string]Value, error)
	Scan(interface{}) error
	Load() interface{}
	Store(interface{})
}

type atomicValue struct {
	atomic.Value
}

func (v *atomicValue) Bool() (bool, error) {
	switch val := v.Load().(type) {
	case bool:
		return val, nil
	case int, int32, int64, float64, string:
		return strconv.ParseBool(fmt.Sprint(val))
	}
	return false, fmt.Errorf("type assert to %v failed", reflect.TypeOf(v.Load()))
}

func (v *atomicValue) Int() (int64, error) {
	switch val := v.Load().(type) {
	case int:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64) //nolint:gomnd
	}
	return 0, fmt.Errorf("type assert to %v failed", reflect.TypeOf(v.Load()))
}

func (v *atomicValue) Slice() ([]Value, error) {
	if vals, ok := v.Load().([]interface{}); ok {
		var slices []Value
		for _, val := range vals {
			a := &atomicValue{}
			a.Store(val)
			slices = append(slices, a)
		}
		return slices, nil
	}
	return nil, fmt.Errorf("type assert to %v failed", reflect.TypeOf(v.Load()))
}

func (v *atomicValue) Map() (map[string]Value, error) {
	if vals, ok := v.Load().(map[string]interface{}); ok {
		m := make(map[string]Value)
		for key, val := range vals {
			a := &atomicValue{}
			a.Store(val)
			m[key] = a
		}
		return m, nil
	}
	return nil, fmt.Errorf("type assert to %v failed", reflect.TypeOf(v.Load()))
}

func (v *atomicValue) Float() (float64, error) {
	switch val := v.Load().(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64) //nolint:gomnd
	}
	return 0.0, fmt.Errorf("type assert to %v failed", reflect.TypeOf(v.Load()))
}

func (v *atomicValue) String() string {
	switch val := v.Load().(type) {
	case string:
		return val
	case bool, int, int32, int64, float64:
		return fmt.Sprint(val)
	case []byte:
		return string(val)
	default:
		if s, ok := val.(fmt.Stringer); ok {
			return s.String()
		}
		return fmt.Sprintf("%+v", val)
	}
}

func (v *atomicValue) Duration() (time.Duration, error) {
	val, err := v.Int()
	if err != nil {
		return 0, err
	}
	return time.Duration(val), nil
}

func (v *atomicValue) Scan(obj interface{}) error {
	data, err := json.Marshal(v.Load())
	if err != nil {
		return err
	}
	if pb, ok := obj.(proto.Message); ok {
		return protojson.Unmarshal(data, pb)
	}
	return json.Unmarshal(data, obj)
}

type emptyValue struct {
	err error
}

func (v emptyValue) Bool() (bool, error)              { return false, v.err }
func (v emptyValue) Int() (int64, error)              { return 0, v.err }
func (v emptyValue) Float() (float64, error)          { return 0.0, v.err }
func (v emptyValue) Duration() (time.Duration, error) { return 0, v.err }
func (v emptyValue) String() string                   { return "" }
func (v emptyValue) Scan(interface{}) error           { return v.err }
func (v emptyValue) Load() interface{}                { return nil }
func (v emptyValue) Store(interface{})                {}
func (v emptyValue) Slice() ([]Value, error)          { return nil, v.err }
func (v emptyValue) Map() (map[string]Value, error)   { return nil, v.err }
