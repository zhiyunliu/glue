package tpl

import (
	"reflect"
	"testing"

	"github.com/zhiyunliu/glue/xdb"
)

func TestIsNil(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := xdb.IsNil(tt.input); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultAnalyze(t *testing.T) {
	var symbols SymbolMap

	type args struct {
		tpl         string
		input       map[string]interface{}
		placeholder xdb.Placeholder
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 *ReplaceItem
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, _ := DefaultAnalyze(symbols, tt.args.tpl, tt.args.input, tt.args.placeholder)
			if got != tt.want {
				t.Errorf("DefaultAnalyze() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("DefaultAnalyze() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestAnalyzeTPLFromCache(t *testing.T) {
	var template *FixedContext
	tmp := NewFixed("test", "?")
	template = tmp.(*FixedContext)
	tests := []struct {
		name       string
		template   SQLTemplate
		tpl        string
		input      map[string]interface{}
		ph         xdb.Placeholder
		wantSql    string
		wantValues []any
	}{
		{
			name: "1.", template: template, tpl: "f=@{f}", input: map[string]any{"f": "1"}, ph: &fixedPlaceHolder{ctx: template}, wantSql: "f=?", wantValues: []any{"1"},
		},
		{
			name: "2.", template: template, tpl: "f=${f}", input: map[string]any{"f": "1"}, ph: &fixedPlaceHolder{ctx: template}, wantSql: "f=1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSql, gotValues, _ := AnalyzeTPLFromCache(tt.template, tt.tpl, tt.input, tt.ph)
			if gotSql != tt.wantSql {
				t.Errorf("AnalyzeTPLFromCache() gotSql = %v, want %v", gotSql, tt.wantSql)
			}
			if !reflect.DeepEqual(gotValues, tt.wantValues) {
				t.Errorf("AnalyzeTPLFromCache() gotValues = %v, want %v", gotValues, tt.wantValues)
			}
		})
	}
}
