package internal

import (
	"fmt"
	"reflect"
	"sync"
)

func typeEncoder(t reflect.Type) encoderFunc {
	if fi, ok := encoderCache.Load(t); ok {
		return fi.(encoderFunc)
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it. This indirect
	// func is only used for recursive types.
	var (
		wg sync.WaitGroup
		f  encoderFunc
	)
	wg.Add(1)
	fi, loaded := encoderCache.LoadOrStore(t, encoderFunc(func(v reflect.Value) any {
		wg.Wait()
		return f(v)
	}))
	if loaded {
		return fi.(encoderFunc)
	}

	// Compute the real encoder and replace the indirect func with it.
	f = newTypeEncoder(t)
	wg.Done()
	encoderCache.Store(t, f)
	return f
}

// newTypeEncoder constructs an encoderFunc for a type.
// The returned encoder only checks CanAddr when allowAddr is true.
func newTypeEncoder(t reflect.Type) encoderFunc {
	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.Float32:
		return float32Encoder
	case reflect.Float64:
		return float64Encoder
	case reflect.String:
		return stringEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Struct:
		return newStructEncoder(t)
	case reflect.Map:
		return newMapEncoder(t)
	case reflect.Slice:
		return newSliceEncoder(t)
	case reflect.Array:
		return newArrayEncoder(t)
	case reflect.Pointer:
		return newPtrEncoder(t)
	default:
		return unsupportedTypeEncoder
	}
}

func unsupportedTypeEncoder(v reflect.Value) any {
	return nil
}

func boolEncoder(v reflect.Value) any {
	return v.Bool()
}

func intEncoder(v reflect.Value) any {
	return v.Int()
}

func uintEncoder(v reflect.Value) any {
	return v.Uint()
}

type floatEncoder int // number of bits

func (bits floatEncoder) encode(v reflect.Value) any {
	return v.Float()
}

var (
	float32Encoder = (floatEncoder(32)).encode
	float64Encoder = (floatEncoder(64)).encode
)

func stringEncoder(v reflect.Value) any {
	return v.String()
}

func interfaceEncoder(v reflect.Value) any {
	return v.Interface()
}

type structEncoder struct {
	fields *structFields
}

func (se structEncoder) encode(v reflect.Value) any {
	if !v.Type().Implements(stringerType) {
		return unsupportedTypeEncoder(v)
	}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	return v.Interface().(fmt.Stringer).String()
}

func newStructEncoder(t reflect.Type) encoderFunc {
	se := structEncoder{fields: cachedTypeFields(t)}
	return se.encode
}

type mapEncoder struct {
	elemEnc encoderFunc
}

func (me mapEncoder) encode(v reflect.Value) any {
	if !v.Type().Implements(stringerType) {
		return unsupportedTypeEncoder(v)
	}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	return v.Interface().(fmt.Stringer).String()
}

func newMapEncoder(t reflect.Type) encoderFunc {
	switch t.Key().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		if !t.Key().Implements(stringerType) {
			return unsupportedTypeEncoder
		}
	}
	me := mapEncoder{elemEnc: typeEncoder(t.Elem())}
	return me.encode
}

func encodeByteSlice(v reflect.Value) any {
	return v.Bytes()
}

// sliceEncoder just wraps an arrayEncoder, checking to make sure the value isn't nil.
type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se sliceEncoder) encode(v reflect.Value) any {
	return se.arrayEnc(v)
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	// Byte slices get special treatment; arrays don't.
	if t.Elem().Kind() == reflect.Uint8 {
		p := reflect.PointerTo(t.Elem())
		if !p.Implements(stringerType) {
			return encodeByteSlice
		}
	}
	enc := sliceEncoder{arrayEnc: newArrayEncoder(t)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

func (ae arrayEncoder) encode(v reflect.Value) any {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return v.Interface()
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	enc := arrayEncoder{elemEnc: typeEncoder(t.Elem())}
	return enc.encode
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe ptrEncoder) encode(v reflect.Value) any {
	if v.IsNil() {
		return nil
	}
	return pe.elemEnc(v.Elem())
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	enc := ptrEncoder{typeEncoder(t.Elem())}
	return enc.encode
}
