package internal

import (
	"reflect"
	"testing"
)

func Test_boolDecoder(t *testing.T) {

	val := struct {
		B1 bool
		B2 *bool
		B3 bool
		B4 *bool
	}{
		B1: false,
	}

	refVal := reflect.ValueOf(&val)
	refVal = refVal.Elem()
	var refTrue *bool = new(bool)
	*refTrue = true

	tests := []struct {
		name    string
		v       reflect.Value
		val     any
		wantErr bool
	}{
		{name: "1.", v: refVal.FieldByName("B1"), val: true, wantErr: false},
		{name: "2.", v: refVal.FieldByName("B2"), val: true, wantErr: false},
		{name: "3.", v: refVal.FieldByName("B3"), val: refTrue, wantErr: false},
		{name: "4.", v: refVal.FieldByName("B4"), val: refTrue, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := boolDecoder(tt.v, tt.val); (err != nil) != tt.wantErr {
				t.Errorf("boolDecoder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_intDecoder(t *testing.T) {
	val := struct {
		B1 int
		B2 *int
		B3 int
		B4 *int
	}{}

	refVal := reflect.ValueOf(&val)
	refVal = refVal.Elem()
	var refTrue *int = new(int)
	*refTrue = 1

	tests := []struct {
		name    string
		v       reflect.Value
		val     any
		wantErr bool
	}{
		{name: "1.", v: refVal.FieldByName("B1"), val: 1, wantErr: false},
		{name: "2.", v: refVal.FieldByName("B2"), val: 1, wantErr: false},
		{name: "3.", v: refVal.FieldByName("B3"), val: refTrue, wantErr: false},
		{name: "4.", v: refVal.FieldByName("B4"), val: refTrue, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := intDecoder(tt.v, tt.val); (err != nil) != tt.wantErr {
				t.Errorf("intDecoder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_stringDecoder(t *testing.T) {
	val := struct {
		B1 string
		B2 *string
		B3 string
		B4 *string
	}{}

	refVal := reflect.ValueOf(&val)
	refVal = refVal.Elem()
	var refTrue *string = new(string)
	*refTrue = "a"
	tests := []struct {
		name    string
		v       reflect.Value
		val     any
		wantErr bool
	}{
		{name: "1.", v: refVal.FieldByName("B1"), val: "a", wantErr: false},
		{name: "2.", v: refVal.FieldByName("B2"), val: "a", wantErr: false},
		{name: "3.", v: refVal.FieldByName("B3"), val: refTrue, wantErr: false},
		{name: "4.", v: refVal.FieldByName("B4"), val: refTrue, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := stringDecoder(tt.v, tt.val); (err != nil) != tt.wantErr {
				t.Errorf("stringDecoder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_floatDecoder_dencode(t *testing.T) {
	val := struct {
		A1 float64
		A2 *float64
		A3 float64
		A4 *float64

		B1 float64
		B2 *float64
		B3 float64
		B4 *float64
	}{}

	refVal := reflect.ValueOf(&val)
	refVal = refVal.Elem()
	var arefTrue *float64 = new(float64)
	*arefTrue = 1.1

	var brefTrue *float64 = new(float64)
	*brefTrue = 2.1

	tests := []struct {
		name    string
		bits    floatDecoder
		v       reflect.Value
		val     any
		wantErr bool
	}{
		{name: "a1.", bits: 32, v: refVal.FieldByName("A1"), val: 1.1, wantErr: false},
		{name: "a2.", bits: 32, v: refVal.FieldByName("A2"), val: 1.1, wantErr: false},
		{name: "a3.", bits: 32, v: refVal.FieldByName("A3"), val: arefTrue, wantErr: false},
		{name: "a4.", bits: 32, v: refVal.FieldByName("A4"), val: arefTrue, wantErr: false},
		{name: "b1.", bits: 64, v: refVal.FieldByName("B1"), val: 2.1, wantErr: false},
		{name: "b2.", bits: 64, v: refVal.FieldByName("B2"), val: 2.1, wantErr: false},
		{name: "b3.", bits: 64, v: refVal.FieldByName("B3"), val: brefTrue, wantErr: false},
		{name: "b4.", bits: 64, v: refVal.FieldByName("B4"), val: brefTrue, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bits.dencode(tt.v, tt.val); (err != nil) != tt.wantErr {
				t.Errorf("floatDecoder.dencode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
