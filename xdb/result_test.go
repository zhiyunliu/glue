package xdb

import (
	"encoding/json"
	"reflect"
	"testing"
)

type TestStruct struct {
	Name string     `json:"name"`
	Raw  RawMessage `json:"raw"`
}

type TestStruct2 struct {
	Name string    `json:"name"`
	Raw  TestInner `json:"raw"`
}

type TestInner struct {
	F1 string     `json:"f1"`
	F2 RawMessage `json:"f2"`
}

func TestRawMessage(t *testing.T) {

	tests := []struct {
		name    string
		Obj     interface{}
		tmp     TestStruct2
		wantRes interface{}
		wantErr bool
	}{
		//		{name: "1.", Obj: &testStruct{Name: "1", Raw: []byte(`{"a":1}`)}, tmp: map[string]any{}, wantRes: map[string]any{"name": "1", "raw": map[string]any{"a": float64(1)}}, wantErr: false},
		//{name: "2.", Obj: Row{"name": "2", "raw": `{"a":1}`}, tmp: TestStruct{}, wantRes: TestStruct{Name: "1", Raw: []byte(`{"a":1}`)}, wantErr: false},
		//{name: "3.", Obj: Row{"name": "3", "raw": map[string]any{"f1": "1", "f2": `{"i":"a","j":2}`}}, tmp: TestStruct2{}, wantRes: TestStruct2{Name: "1", Raw: TestInner{F1: "1", F2: []byte(`{"i":"a","j":2}`)}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := json.Marshal(tt.Obj)
			if err != nil {
				t.Error(err)
				return
			}
			err = json.Unmarshal(bytes, &tt.tmp)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(tt.tmp, tt.wantRes) {
				t.Errorf("TestRawMessage = %v, want %v", tt.tmp, tt.wantRes)
			}
		})
	}
}
