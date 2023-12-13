package internal

import (
	"database/sql"
	"reflect"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xtypes"
)

func typeDencoder(t reflect.Type) dencoderFunc {
	if fi, ok := dencoderCache.Load(t); ok {
		return fi.(dencoderFunc)
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it. This indirect
	// func is only used for recursive types.
	var (
		wg sync.WaitGroup
		f  dencoderFunc
	)
	wg.Add(1)
	fi, loaded := dencoderCache.LoadOrStore(t, dencoderFunc(func(v reflect.Value, val any) error {
		wg.Wait()
		return f(v, val)
	}))
	if loaded {
		return fi.(dencoderFunc)
	}

	// Compute the real encoder and replace the indirect func with it.
	f = newTypeDencoder(t)
	wg.Done()
	dencoderCache.Store(t, f)
	return f
}

func newTypeDencoder(t reflect.Type) dencoderFunc {
	switch t.Kind() {
	case reflect.Bool:
		return boolDecoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intDecoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintDecoder
	case reflect.Float32:
		return float32Decoder
	case reflect.Float64:
		return float64Decoder
	case reflect.String:
		return stringDecoder
	case reflect.Struct:
		return structDecoder
	case reflect.Map:
		return mapDecoder
	case reflect.Interface:
		return interfaceDecoder
	case reflect.Slice:
		return newSliceDecoder(t)
	case reflect.Array:
		return newArrayDecoder(t)
	default:
		return defaultDecoder

		// case reflect.Pointer:
		// 	return newPtrDecoder(t)
		// default:
		// 	return unsupportedTypeDecoder
	}
}

// func unsupportedTypeDecoder(v reflect.Value, val any) error {
// 	if v.Type().NumMethod() > 0 && v.CanInterface() {
// 		if u, ok := v.Interface().(sql.Scanner); ok {
// 			return u.Scan(val)
// 		}
// 	}
// 	return nil
// }

func boolDecoder(v reflect.Value, val any) error {
	rv := reflect.ValueOf(val)
	if rv.IsZero() {
		return nil
	}

	if rv.Kind() == reflect.Pointer {
		val = rv.Elem().Interface()
	}

	tmpv := xtypes.GetBool(val)
	if v.Kind() == reflect.Pointer {
		tmprv := reflect.New(v.Type().Elem())
		tmprv.Elem().SetBool(tmpv)
		v.Set(tmprv)
		return nil
	}
	v.SetBool(tmpv)
	return nil
}

func intDecoder(v reflect.Value, val any) error {
	rv := reflect.ValueOf(val)
	if rv.IsZero() {
		return nil
	}
	intval, err := xtypes.GetInt64(val)
	if err != nil {
		return err
	}
	if v.Kind() == reflect.Pointer {
		tmprv := reflect.New(v.Type().Elem())
		tmprv.Elem().SetInt(intval)
		v.Set(tmprv)
		return nil
	}
	v.SetInt(intval)
	return nil
}

func uintDecoder(v reflect.Value, val any) error {
	rv := reflect.ValueOf(val)
	if rv.IsZero() {
		return nil
	}
	int64val, err := xtypes.GetInt64(val)
	if err != nil {
		return err
	}
	if v.Kind() == reflect.Pointer {
		tmprv := reflect.New(v.Type().Elem())
		tmprv.Elem().SetUint(uint64(int64val))
		v.Set(tmprv)
		return nil
	}
	v.SetUint(uint64(int64val))
	return nil

}

type floatDecoder int // number of bits

func (bits floatDecoder) dencode(v reflect.Value, val any) error {
	rv := reflect.ValueOf(val)
	if rv.IsZero() {
		return nil
	}
	f64val, err := xtypes.GetFloat64(val)
	if err != nil {
		return err
	}
	if v.Kind() == reflect.Pointer {
		tmprv := reflect.New(v.Type().Elem())
		tmprv.Elem().SetFloat(f64val)
		v.Set(tmprv)
		return nil
	}
	v.SetFloat(f64val)
	return nil
}

var (
	float32Decoder = (floatDecoder(32)).dencode
	float64Decoder = (floatDecoder(64)).dencode
)

func stringDecoder(v reflect.Value, val any) error {
	rv := reflect.ValueOf(val)
	if rv.IsZero() {
		return nil
	}

	if rv.Kind() == reflect.Pointer {
		val = rv.Elem().Interface()
	}

	if v.Kind() == reflect.Pointer {
		tmprv := reflect.New(v.Type().Elem())
		tmprv.Elem().SetString(xtypes.GetString(val))
		v.Set(tmprv)
		return nil
	}
	v.SetString(xtypes.GetString(val))
	return nil
}

func interfaceDecoder(v reflect.Value, val any) error {
	if v.CanSet() {
		v.Set(reflect.ValueOf(val))
	}
	return nil
}

func mapScanDecoder(v reflect.Value, val any) error {
	if val == nil {
		return nil
	}

	if scanner, ok := v.Interface().(xdb.MapScanner); ok {
		return scanner.MapScan(val)
	}
	if v.CanAddr() {
		if scanner, ok := v.Addr().Interface().(xdb.MapScanner); ok {
			return scanner.MapScan(val)
		}
	}
	return nil

}

func mapDecoder(v reflect.Value, val any) error {
	if v.IsNil() {
		if v.Kind() == reflect.Pointer {
			ftyp := v.Type().Elem()
			rv1 := reflect.New(ftyp)
			mapval := reflect.MakeMap(reflect.MapOf(ftyp.Key(), ftyp.Elem()))
			rv1.Elem().Set(mapval)
			v.Set(rv1)
		} else {
			v.Set(reflect.MakeMap(v.Type()))
		}
	}
	return mapScanDecoder(v, val)
}

func dencodeByteSlice(v reflect.Value, val any) error {
	rv := reflect.ValueOf(val)
	if rv.IsZero() {
		return nil
	}
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.IsNil() {
		return nil
	}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	v.Set(rv)
	return nil
}

// sliceDecoder just wraps an arrayDecoder, checking to make sure the value isn't nil.
type sliceDecoder struct {
	arrayEnc dencoderFunc
}

func (se sliceDecoder) dencode(v reflect.Value, val any) error {
	return se.arrayEnc(v, val)
}

func newSliceDecoder(t reflect.Type) dencoderFunc {
	// Byte slices get special treatment; arrays don't.
	if t.Elem().Kind() == reflect.Uint8 {
		p := reflect.PointerTo(t.Elem())
		if !p.Implements(scannerType) {
			return dencodeByteSlice
		}
	}
	enc := sliceDecoder{arrayEnc: newArrayDecoder(t)}
	return enc.dencode
}

type arrayDecoder struct {
	elemEnc dencoderFunc
}

func (ae arrayDecoder) dencode(v reflect.Value, val any) error {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.CanConvert(xmapsType) {
		return mapScanDecoder(v, val)
	}
	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Pointer && rv.IsNil() {
		return nil
	}

	v.Set(rv)
	return nil
}

func newArrayDecoder(t reflect.Type) dencoderFunc {
	enc := arrayDecoder{elemEnc: typeDencoder(t.Elem())}
	return enc.dencode
}

// type ptrDecoder struct {
// 	elemEnc dencoderFunc
// }

// func (pe ptrDecoder) encode(v reflect.Value, val any) error {
// 	return pe.elemEnc(v.Elem(), val)

// }

// func newPtrDecoder(t reflect.Type) dencoderFunc {
// 	enc := ptrDecoder{typeDencoder(t.Elem())}
// 	return enc.encode
// }

func structDecoder(v reflect.Value, val any) error {
	refVal := reflect.ValueOf(val)
	if refVal.IsZero() {
		return nil
	}

	if v.Kind() == reflect.Pointer && v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}

	ftype := v.Type()
	switch {
	case ftype == refVal.Type():
		v.Set(refVal)
		return nil
	case ftype.Kind() == reflect.Pointer && ftype.Elem() == refVal.Type():
		v.Elem().Set(refVal)
		return nil

	case refVal.Kind() == reflect.Pointer && ftype == refVal.Elem().Type():
		v.Set(refVal.Elem())
		return nil
	}

	if scanner, ok := v.Interface().(sql.Scanner); ok {
		return scanner.Scan(val)
	}
	if v.CanAddr() {
		if scanner, ok := v.Addr().Interface().(sql.Scanner); ok {
			return scanner.Scan(val)
		}
	}

	return nil
}

func defaultDecoder(v reflect.Value, val any) error {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Map:
		return mapDecoder(v, val)
	default:
	}
	return nil
}
