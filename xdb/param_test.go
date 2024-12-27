package xdb

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"
)

func TestIsNil(t *testing.T) {

	type Str string

	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{name: "0.", input: nil, want: true},
		{name: "1.", input: (*int)(nil), want: true},
		{name: "2.", input: 1, want: false},
		{name: "3.", input: 1.0, want: false},
		{name: "4.", input: struct{ a int }{a: 1}, want: false},
		{name: "5.", input: struct{ a *int }{}, want: false},
		{name: "6.", input: map[string]interface{}{}, want: false},
		{name: "7.", input: []string{}, want: false},
		{name: "8.", input: []string{"a"}, want: false},
		{name: "9.", input: []int{}, want: false},
		{name: "10.", input: []int{1}, want: false},
		{name: "11.", input: sql.Named("a", ""), want: true},
		{name: "12.", input: sql.Named("a", "a"), want: false},
		{name: "12.", input: &sql.NamedArg{Name: "a", Value: nil}, want: true},

		{name: "12.", input: Str("aaa"), want: false},
		{name: "12.", input: Str(""), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultIsNil(tt.input); got != tt.want {
				t.Errorf("DefaultIsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试 GetVal 方法
func TestGetVal(t *testing.T) {
	tests := []struct {
		name   string
		param  DBParam
		input  string
		output interface{}
		err    MissError
	}{
		{
			name:   "Missing parameter",
			param:  DBParam{},
			input:  "missing",
			output: nil,
			err:    NewMissParamError("missing", nil),
		},
		{
			name: "NamedArg",
			param: DBParam{
				"arg1": sql.Named("arg1", "value1"),
			},
			input:  "arg1",
			output: sql.Named("arg1", "value1"),
			err:    nil,
		},
		{
			name: "Driver Valuer",
			param: DBParam{
				"arg2": &customValuer{value: "value2"},
			},
			input:  "arg2",
			output: "value2",
			err:    nil,
		},
		{
			name: "Driver Valuer Mill",
			param: DBParam{
				"arg2": &customValuer{err: fmt.Errorf("error")},
			},
			input:  "arg2",
			output: nil,
			err:    NewMissParamError("arg2", nil),
		},
		{
			name: "Time",
			param: DBParam{
				"arg3": time.Now(),
			},
			input:  "arg3",
			output: time.Now().Format(DateFormat),
			err:    nil,
		},
		{
			name: "Nil time",
			param: DBParam{
				"arg4": (*time.Time)(nil),
			},
			input:  "arg4",
			output: nil,
			err:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.param.GetVal(tt.input)
			if err != nil && err.Name() != tt.err.Name() {
				t.Errorf("expected error %v, got %v", tt.err, err)
			}
			if fmt.Sprintf("%v", val) != fmt.Sprintf("%v", tt.output) {
				t.Errorf("expected output %v, got %v", tt.output, val)
			}
		})
	}
}

// 自定义 Valuer 实现
type customValuer struct {
	value string
	err   error
}

func (c *customValuer) Value() (driver.Value, error) {
	return c.value, c.err
}
