package internal

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/zhiyunliu/golibs/xtypes"
)

type dbparam struct {
	A string
}

func (p dbparam) ToDbParam() map[string]any {
	return map[string]any{"a": p.A}
}

type jsonStructParam struct {
	A    string             `json:"a"`
	B    int                `json:"b"`
	C    []int              `json:"c"`
	D    *[]int             `json:"d"`
	E    []string           `json:"e"`
	Dec  xtypes.Decimal     `json:"dec"`
	Dec2 *xtypes.Decimal    `json:"dec2"`
	Map  map[string]string  `json:"map"`
	Map2 *map[string]string `json:"map2"`
	Map3 StrMap             `json:"map3"`
	Map4 *StrMap            `json:"map4"`
	Obj  dbstructParam      `json:"obj"`
	Obj2 *dbstructParam     `json:"obj2"`
	Obj3 dbstructParamStr   `json:"obj3"`
	Obj4 *dbstructParamStr  `json:"obj4"`
}

type dbstructJson struct {
	A string `json:"a"`
}

type dbstructParam struct {
	A string `db:"a"`
}

type dbstructParamStr struct {
	A string `db:"a"`
}

func (p dbstructParamStr) String() string {
	return p.A
}

type StrMap map[string]string

func (m StrMap) String() string {
	mapBytes, _ := json.Marshal(m)
	return string(mapBytes)
}

func Test_analyzeParamFields(t *testing.T) {
	decVal := xtypes.NewDecimalFromInt(10)
	mapVal := map[string]string{"m": "m"}
	var smap StrMap = mapVal
	tests := []struct {
		name       string
		input      any
		wantParams xtypes.XMap
		wantErr    bool
	}{
		{name: "1.", input: jsonStructParam{
			A:    "1",
			B:    2,
			C:    []int{1, 2},
			D:    &[]int{1, 2},
			E:    []string{"a", "b"},
			Dec:  decVal,
			Dec2: &decVal,
			Map:  mapVal,
			Map2: &mapVal,
			Map3: smap,
			Map4: &smap,
			Obj:  dbstructParam{A: "obj"},
			Obj2: &dbstructParam{A: "obj2"},
			Obj3: dbstructParamStr{A: "obj3"},
			Obj4: &dbstructParamStr{A: "obj4"},
		}, wantParams: map[string]any{
			"a":    "1",
			"b":    int64(2),
			"c":    []int{1, 2},
			"d":    []int{1, 2},
			"e":    []string{"a", "b"},
			"dec":  "10",
			"dec2": "10",
			"map":  nil,
			"map2": nil,
			"map3": `{"m":"m"}`,
			"map4": `{"m":"m"}`,
			"obj":  nil,
			"obj2": nil,
			"obj3": "obj3",
			"obj4": "obj4",
		}, wantErr: false},

		{name: "2.", input: &jsonStructParam{
			A:    "1",
			B:    2,
			C:    []int{1, 2},
			D:    &[]int{1, 2},
			E:    []string{"a", "b"},
			Dec:  decVal,
			Dec2: &decVal,
			Map:  mapVal,
			Map2: &mapVal,
			Map3: smap,
			Map4: &smap,
			Obj:  dbstructParam{A: "obj"},
			Obj2: &dbstructParam{A: "obj2"},
			Obj3: dbstructParamStr{A: "obj3"},
			Obj4: &dbstructParamStr{A: "obj4"},
		}, wantParams: map[string]any{
			"a":    "1",
			"b":    int64(2),
			"c":    []int{1, 2},
			"d":    []int{1, 2},
			"e":    []string{"a", "b"},
			"dec":  "10",
			"dec2": "10",
			"map":  nil,
			"map2": nil,
			"map3": `{"m":"m"}`,
			"map4": `{"m":"m"}`,
			"obj":  nil,
			"obj2": nil,
			"obj3": "obj3",
			"obj4": "obj4",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotParams, err := analyzeParamFields(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("analyzeParamFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(gotParams, tt.wantParams) {
				t.Errorf("analyzeParamFields() = %v, want %v", gotParams, tt.wantParams)
			}

			for k, v := range gotParams {
				if !reflect.DeepEqual(v, tt.wantParams[k]) {
					t.Errorf("analyzeParamFields() %s = %v, want %v", k, v, tt.wantParams[k])
				}
			}
		})
	}
}

func TestResolveParams(t *testing.T) {

	tests := []struct {
		name       string
		input      any
		wantParams xtypes.XMap
		wantErr    bool
	}{
		{name: "1.", input: map[string]any{"a": 1, "b": 2}, wantParams: map[string]any{"a": 1, "b": 2}, wantErr: false},
		{name: "2.", input: xtypes.XMap{"a": 1, "b": 2}, wantParams: map[string]any{"a": 1, "b": 2}, wantErr: false},
		{name: "3.", input: xtypes.SMap{"a": "1", "b": "2"}, wantParams: map[string]any{"a": "1", "b": "2"}, wantErr: false},
		{name: "4.", input: map[string]string{"a": "1", "b": "2"}, wantParams: map[string]any{"a": "1", "b": "2"}, wantErr: false},
		{name: "5.", input: dbparam{A: "1"}, wantParams: map[string]any{"a": "1"}, wantErr: false},
		{name: "6.", input: &dbparam{A: "1"}, wantParams: map[string]any{"a": "1"}, wantErr: false},
		{name: "7.", input: dbstructParam{A: "1"}, wantParams: map[string]any{"a": "1"}, wantErr: false},
		{name: "8.", input: &dbstructParam{A: "1"}, wantParams: map[string]any{"a": "1"}, wantErr: false},
		{name: "9.", input: dbstructJson{A: "1"}, wantParams: map[string]any{"a": "1"}, wantErr: false},
		{name: "10.", input: &dbstructJson{A: "1"}, wantParams: map[string]any{"a": "1"}, wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotParams, err := ResolveParams(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotParams, tt.wantParams) {
				t.Errorf("ResolveParams() = %v, want %v", gotParams, tt.wantParams)
			}
		})
	}
}
