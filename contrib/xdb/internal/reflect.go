package internal

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	fieldCache   sync.Map
	stringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
)

type encoderFunc func(v reflect.Value) any

type structFields struct {
	list []field
}

type field struct {
	name      string
	index     int
	tag       bool
	typ       reflect.Type
	omitEmpty bool
	encoder   encoderFunc
}

func cachedTypeFields(t reflect.Type) structFields {
	if f, ok := fieldCache.Load(t); ok {
		return f.(structFields)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.(structFields)
}

// typeFields returns a list of fields that JSON should recognize for the given type.
// The algorithm is breadth-first search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
func typeFields(t reflect.Type) structFields {
	// Anonymous fields to explore at the current level and the next.
	current := []field{}
	next := []field{{typ: t}}

	// Count of queued names for current level and the next.
	var count, nextCount map[reflect.Type]int

	// Types already visited at an earlier level.
	visited := map[reflect.Type]bool{}

	// Fields found.
	var fields []field

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// Scan f.typ for fields to include.
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Pointer {
						t = t.Elem()
					}
					if !sf.IsExported() && t.Kind() != reflect.Struct {
						// Ignore embedded fields of unexported non-struct types.
						continue
					}
					// Do not ignore embedded fields of unexported struct types
					// since they may have exported fields.
				} else if !sf.IsExported() {
					// Ignore unexported non-embedded fields.
					continue
				}
				tag := sf.Tag.Get("db")
				if tag == "" {
					tag = sf.Tag.Get("json")
					if tag == "-" {
						continue
					}
				}

				name, opts := parseTag(tag)
				if !isValidTag(name) {
					name = ""
				}

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Pointer {
					// Follow pointer.
					ft = ft.Elem()
				}

				// Record found field and index sequence.
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					field := field{
						name:      name,
						index:     i,
						tag:       tagged,
						typ:       ft,
						omitEmpty: opts.Contains("omitempty"),
					}

					fields = append(fields, field)
					if count[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}
			}
		}
	}

	for i := range fields {
		f := &fields[i]
		f.encoder = typeEncoder(f.typ)
	}

	return structFields{list: fields}
}

var encoderCache sync.Map // map[reflect.Type]encoderFunc

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
	fields structFields
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
