package internal

import (
	"reflect"
	"sync"
)

type structFields struct {
	list      []field
	exactName map[string]*field
}

type field struct {
	name         string
	fieldName    string
	index        int
	typ          reflect.Type
	orgtyp       reflect.Type
	indirectType reflect.Type
	tag          reflect.StructTag
	omitEmpty    bool
	encoder      encoderFunc
	dencoder     dencoderFunc
	Schema       *Schema
	NewValuePool FieldNewValuePool
}

type Schema struct {
	Name         string
	ModelType    reflect.Type
	Fields       []*field
	FieldsByName map[string]*field
	initialized  chan struct{}
	err          error
}

func (schema *Schema) ParseField(fieldStruct reflect.StructField) *field {

	field := &field{
		fieldName:    fieldStruct.Name,
		typ:          fieldStruct.Type,
		indirectType: fieldStruct.Type,
		tag:          fieldStruct.Tag,
		Schema:       schema,
	}

	for field.indirectType.Kind() == reflect.Ptr {
		field.indirectType = field.indirectType.Elem()
	}
	field.setupNewValuePool()
	return field
}

func (field *field) setupNewValuePool() {
	if field.NewValuePool == nil {
		field.NewValuePool = poolInitializer(reflect.PtrTo(field.indirectType))
	}
}

var (
	normalPool      sync.Map
	poolInitializer = func(reflectType reflect.Type) FieldNewValuePool {
		v, _ := normalPool.LoadOrStore(reflectType, &sync.Pool{
			New: func() interface{} {
				return reflect.New(reflectType).Interface()
			},
		})
		return v.(FieldNewValuePool)
	}
)

type FieldNewValuePool interface {
	Get() interface{}
	Put(interface{})
}
