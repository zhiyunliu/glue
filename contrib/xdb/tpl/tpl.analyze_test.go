package tpl

import (
	"reflect"
	"testing"
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
			if got := IsNil(tt.input); got != tt.want {
				t.Errorf("IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultAnalyze(t *testing.T) {
	var symbols Symbols

	type args struct {
		tpl         string
		input       map[string]interface{}
		placeholder Placeholder
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
	type args struct {
		template SQLTemplate
		tpl      string
		input    map[string]interface{}
		ph       Placeholder
	}
	tests := []struct {
		name       string
		args       args
		wantSql    string
		wantValues []any
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSql, gotValues, _ := AnalyzeTPLFromCache(tt.args.template, tt.args.tpl, tt.args.input, tt.args.ph)
			if gotSql != tt.wantSql {
				t.Errorf("AnalyzeTPLFromCache() gotSql = %v, want %v", gotSql, tt.wantSql)
			}
			if !reflect.DeepEqual(gotValues, tt.wantValues) {
				t.Errorf("AnalyzeTPLFromCache() gotValues = %v, want %v", gotValues, tt.wantValues)
			}
		})
	}
}
