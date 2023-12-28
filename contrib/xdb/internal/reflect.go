package internal

import (
	"database/sql"
	"fmt"
	"reflect"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var (
	fieldCache        sync.Map
	encoderCache      sync.Map
	dencoderCache     sync.Map
	stringerType      = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	scannerType       = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	xmapsType         = reflect.TypeOf((*xdb.Rows)(nil)).Elem()
	structScannerType = reflect.TypeOf((*xdb.StructScanner)(nil)).Elem()
)

type encoderFunc func(v reflect.Value) any
type dencoderFunc func(reflect.Value, any) error

func cachedTypeFields(t reflect.Type) *structFields {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if f, ok := fieldCache.Load(t); ok {
		return f.(*structFields)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.(*structFields)
}

// typeFields returns a list of fields that JSON should recognize for the given type.
// The algorithm is breadth-first search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
func typeFields(t reflect.Type) *structFields {

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
					if name == "" {
						name = sf.Name
					}
					field := field{
						name:      name,
						fieldName: sf.Name,
						index:     i,
						typ:       ft,
						orgtyp:    sf.Type,
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
	exactName := make(map[string]*field, len(fields))
	for i := range fields {
		f := &fields[i]
		exactName[f.name] = &fields[i]
		f.encoder = typeEncoder(f.typ)
		f.dencoder = typeDencoder(f.typ)
	}

	return &structFields{list: fields, exactName: exactName}
}

// func setReflectVal(field *field) {

// 	// ReflectValueOf returns field's reflect value
// 	fieldIndex := field.index
// 	switch {
// 	case len(field.StructField.Index) == 1 && fieldIndex > 0:
// 		field.ReflectValueOf = func(value reflect.Value) reflect.Value {
// 			return reflect.Indirect(value).Field(fieldIndex)
// 		}
// 	default:
// 		field.ReflectValueOf = func(v reflect.Value) reflect.Value {
// 			v = reflect.Indirect(v)
// 			for idx, fieldIdx := range field.StructField.Index {
// 				if fieldIdx >= 0 {
// 					v = v.Field(fieldIdx)
// 				} else {
// 					v = v.Field(-fieldIdx - 1)

// 					if v.IsNil() {
// 						v.Set(reflect.New(v.Type().Elem()))
// 					}

// 					if idx < len(field.StructField.Index)-1 {
// 						v = v.Elem()
// 					}
// 				}
// 			}
// 			return v
// 		}
// 	}

// }
