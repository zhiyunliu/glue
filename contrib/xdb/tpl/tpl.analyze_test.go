package tpl

import (
	"reflect"
	"testing"

	"github.com/zhiyunliu/glue/contrib/xdb/expression"
	"github.com/zhiyunliu/glue/xdb"
)

func Test_NewFixed_AnalyzeTPLFromCache(t *testing.T) {

	fixedtemplate := NewFixed("test",
		"?",
		NewDefaultTemplateMatcher(expression.DefaultExpressionMatchers...),
		NewDefaultStmtDbTypeProcessor(),
	)

	tests := []struct {
		name       string
		template   xdb.SQLTemplate
		tpl        string
		input      map[string]interface{}
		wantSql    string
		wantValues []any
		wantErr    bool
	}{
		{name: "a-1-1.", tpl: "f=@{f}", input: map[string]any{"f": "1"}, wantSql: "f=?", wantValues: []any{"1"}},
		{name: "a-1-2.", tpl: "f=${f}", input: map[string]any{"f": "1"}, wantSql: "f=1", wantValues: nil},
		{name: "a-1-3.", tpl: "&{f}", input: map[string]any{"f": "2"}, wantSql: "and f=?", wantValues: []any{"2"}},
		{name: "a-1-4.", tpl: "&{t.f}", input: map[string]any{"f": "3"}, wantSql: "and t.f=?", wantValues: []any{"3"}},
		{name: "a-1-5.", tpl: "|{f}", input: map[string]any{"f": "4"}, wantSql: "or f=?", wantValues: []any{"4"}},
		{name: "a-1-6.", tpl: "|{t.f}", input: map[string]any{"f": "5"}, wantSql: "or t.f=?", wantValues: []any{"5"}},

		{name: "a-2-1.", tpl: "f=@{f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "f=? and b=?", wantValues: []any{"1", "1"}},
		{name: "a-2-2.", tpl: "f=${f} and b=@{f}", input: map[string]any{"f": "2"}, wantSql: "f=2 and b=?", wantValues: []any{"2"}},
		{name: "a-2-3.", tpl: "&{f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "and f=? and b=?", wantValues: []any{"1", "1"}},
		{name: "a-2-4.", tpl: "&{t.f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "and t.f=? and b=?", wantValues: []any{"1", "1"}},
		{name: "a-2-5.", tpl: "|{f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "or f=? and b=?", wantValues: []any{"1", "1"}},
		{name: "a-2-6.", tpl: "|{t.f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "or t.f=? and b=?", wantValues: []any{"1", "1"}},

		{name: "b-1-1.", tpl: "&{like f}", input: map[string]any{"f": "3"}, wantSql: "and f like ?", wantValues: []any{"3"}},
		{name: "b-1-2.", tpl: "&{like t.f}", input: map[string]any{"f": "4"}, wantSql: "and t.f like ?", wantValues: []any{"4"}},
		{name: "b-1-3.", tpl: "&{like %f}", input: map[string]any{"f": "5"}, wantSql: "and f like '%'+?", wantValues: []any{"5"}},
		{name: "b-1-4.", tpl: "&{like %t.f}", input: map[string]any{"f": "6"}, wantSql: "and t.f like '%'+?", wantValues: []any{"6"}},
		{name: "b-1-5.", tpl: "&{like %f%}", input: map[string]any{"f": "7"}, wantSql: "and f like '%'+?+'%'", wantValues: []any{"7"}},
		{name: "b-1-6.", tpl: "&{like %t.f%}", input: map[string]any{"f": "8"}, wantSql: "and t.f like '%'+?+'%'", wantValues: []any{"8"}},

		{name: "b-2-1.", tpl: "|{like f} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or f like ? and t.f=?", wantValues: []any{"1", "1"}},
		{name: "b-2-2.", tpl: "|{like t.f} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or t.f like ? and t.f=?", wantValues: []any{"1", "1"}},
		{name: "b-2-3.", tpl: "|{like %f} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or f like '%'+? and t.f=?", wantValues: []any{"1", "1"}},
		{name: "b-2-4.", tpl: "|{like %t.f} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or t.f like '%'+? and t.f=?", wantValues: []any{"1", "1"}},
		{name: "b-2-5.", tpl: "|{like %f%} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or f like '%'+?+'%' and t.f=?", wantValues: []any{"1", "1"}},
		{name: "b-2-6.", tpl: "|{like %t.f%} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or t.f like '%'+?+'%' and t.f=?", wantValues: []any{"1", "1"}},

		{name: "c-1-1.", tpl: "&{in f}", input: map[string]any{"f": []string{"1"}}, wantSql: "and f in ('1')", wantValues: nil},
		{name: "c-1-2.", tpl: "&{in t.f}", input: map[string]any{"f": []string{"1", "2"}}, wantSql: "and t.f in ('1','2')", wantValues: nil},
		{name: "c-1-3.", tpl: "&{in t.f}", input: map[string]any{"f": []string{"1", "'"}}, wantSql: "and t.f in ('1','''')", wantValues: nil},
		{name: "c-1-4.", tpl: "&{in t.f}", input: map[string]any{"f": []string{"1", "2", "'--"}}, wantSql: "and t.f in ('1','2','''--')", wantValues: nil},
		{name: "c-1-5.", tpl: "&{in f}", input: map[string]any{"f": []int{1, 2}}, wantSql: "and f in (1,2)", wantValues: nil},
		{name: "c-1-6.", tpl: "&{in t.f}", input: map[string]any{"f": []int{1, 2}}, wantSql: "and t.f in (1,2)", wantValues: nil},

		{name: "c-2-1.", tpl: "|{filed in f}", input: map[string]any{"f": []string{"1"}}, wantSql: "or filed in ('1')", wantValues: nil},
		{name: "c-2-2.", tpl: "|{t.field in f}", input: map[string]any{"f": []string{"1"}}, wantSql: "or t.field in ('1')", wantValues: nil},
		{name: "c-2-3.", tpl: "|{in f}", input: map[string]any{"f": []int{1, 2}}, wantSql: "or f in (1,2)", wantValues: nil},
		{name: "c-2-4.", tpl: "|{in t.f}", input: map[string]any{"f": []int{1, 2}}, wantSql: "or t.f in (1,2)", wantValues: nil},

		{name: "d-1-1.", tpl: "&{> f}", input: map[string]any{"f": "1"}, wantSql: "and f>?", wantValues: []any{"1"}},
		{name: "d-1-2.", tpl: "&{>= t.f}", input: map[string]any{"f": "2"}, wantSql: "and t.f>=?", wantValues: []any{"2"}},
		{name: "d-1-3.", tpl: "&{< f}", input: map[string]any{"f": "3"}, wantSql: "and f<?", wantValues: []any{"3"}},
		{name: "d-1-4.", tpl: "&{<= t.f}", input: map[string]any{"f": "4"}, wantSql: "and t.f<=?", wantValues: []any{"4"}},
		{name: "d-1-5.", tpl: "&{= t.f}", input: map[string]any{"f": "5"}, wantSql: "and t.f=?", wantValues: []any{"5"}},

		{name: "d-2-1.", tpl: "&{field > f}", input: map[string]any{"f": "1"}, wantSql: "and field>?", wantValues: []any{"1"}},
		{name: "d-2-2.", tpl: "&{t.field >= f}", input: map[string]any{"f": "2"}, wantSql: "and t.field>=?", wantValues: []any{"2"}},
		{name: "d-2-3.", tpl: "&{field < f}", input: map[string]any{"f": "3"}, wantSql: "and field<?", wantValues: []any{"3"}},
		{name: "d-2-4.", tpl: "&{t.field <= f}", input: map[string]any{"f": "4"}, wantSql: "and t.field<=?", wantValues: []any{"4"}},
		{name: "d-2-5.", tpl: "&{t.field = f}", input: map[string]any{"f": "5"}, wantSql: "and t.field=?", wantValues: []any{"5"}},
	}

	for _, tt := range tests {
		tt.template = fixedtemplate
		t.Run(tt.name, func(t *testing.T) {
			gotSql, gotValues, err := AnalyzeTPLFromCache(tt.template, tt.tpl, tt.input)
			if tt.wantErr != (err != nil) {
				t.Errorf("AnalyzeTPLFromCache() wantErr = %v, want %v", err, tt.wantErr)
			}
			if gotSql != tt.wantSql {
				t.Errorf("AnalyzeTPLFromCache() gotSql = %v, want %v", gotSql, tt.wantSql)
			}
			if !reflect.DeepEqual(gotValues, tt.wantValues) {
				t.Errorf("AnalyzeTPLFromCache() gotValues = %v, want %v", gotValues, tt.wantValues)
			}
		})
	}

}

func Test_NewSeq_AnalyzeTPLFromCache(t *testing.T) {

	fixedtemplate := NewSeq("test",
		":",
		NewDefaultTemplateMatcher(expression.DefaultExpressionMatchers...),
		NewDefaultStmtDbTypeProcessor(),
	)

	tests := []struct {
		name       string
		template   xdb.SQLTemplate
		tpl        string
		input      map[string]interface{}
		ph         xdb.Placeholder
		wantSql    string
		wantValues []any
		wantErr    bool
	}{
		{name: "a-1-1.", tpl: "f=@{f}", input: map[string]any{"f": "1"}, wantSql: "f=:1", wantValues: []any{"1"}},
		{name: "a-1-2.", tpl: "f=${f}", input: map[string]any{"f": "1"}, wantSql: "f=1", wantValues: nil},
		{name: "a-1-3.", tpl: "&{f}", input: map[string]any{"f": "1"}, wantSql: "and f=:1", wantValues: []any{"1"}},
		{name: "a-1-4.", tpl: "&{t.f}", input: map[string]any{"f": "1"}, wantSql: "and t.f=:1", wantValues: []any{"1"}},
		{name: "a-1-5.", tpl: "|{f}", input: map[string]any{"f": "1"}, wantSql: "or f=:1", wantValues: []any{"1"}},
		{name: "a-1-6.", tpl: "|{t.f}", input: map[string]any{"f": "1"}, wantSql: "or t.f=:1", wantValues: []any{"1"}},

		{name: "a-2-1.", tpl: "f=@{f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "f=:1 and b=:2", wantValues: []any{"1", "1"}},
		{name: "a-2-2.", tpl: "f=${f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "f=1 and b=:1", wantValues: []any{"1"}},
		{name: "a-2-3.", tpl: "&{f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "and f=:1 and b=:2", wantValues: []any{"1", "1"}},
		{name: "a-2-4.", tpl: "&{t.f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "and t.f=:1 and b=:2", wantValues: []any{"1", "1"}},
		{name: "a-2-5.", tpl: "|{f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "or f=:1 and b=:2", wantValues: []any{"1", "1"}},
		{name: "a-2-6.", tpl: "|{t.f} and b=@{f}", input: map[string]any{"f": "1"}, wantSql: "or t.f=:1 and b=:2", wantValues: []any{"1", "1"}},

		{name: "b-1-1.", tpl: "&{like f}", input: map[string]any{"f": "1"}, wantSql: "and f like :1", wantValues: []any{"1"}},
		{name: "b-1-2.", tpl: "&{like t.f}", input: map[string]any{"f": "1"}, wantSql: "and t.f like :1", wantValues: []any{"1"}},
		{name: "b-1-3.", tpl: "&{like %f}", input: map[string]any{"f": "1"}, wantSql: "and f like '%'+:1", wantValues: []any{"1"}},
		{name: "b-1-4.", tpl: "&{like %t.f}", input: map[string]any{"f": "1"}, wantSql: "and t.f like '%'+:1", wantValues: []any{"1"}},
		{name: "b-1-5.", tpl: "&{like %f%}", input: map[string]any{"f": "1"}, wantSql: "and f like '%'+:1+'%'", wantValues: []any{"1"}},
		{name: "b-1-6.", tpl: "&{like %t.f%}", input: map[string]any{"f": "1"}, wantSql: "and t.f like '%'+:1+'%'", wantValues: []any{"1"}},

		{name: "b-2-1.", tpl: "|{like f} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or f like :1 and t.f=:2", wantValues: []any{"1", "1"}},
		{name: "b-2-2.", tpl: "|{like t.f} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or t.f like :1 and t.f=:2", wantValues: []any{"1", "1"}},
		{name: "b-2-3.", tpl: "|{like %f} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or f like '%'+:1 and t.f=:2", wantValues: []any{"1", "1"}},
		{name: "b-2-4.", tpl: "|{like %t.f} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or t.f like '%'+:1 and t.f=:2", wantValues: []any{"1", "1"}},
		{name: "b-2-5.", tpl: "|{like %f%} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or f like '%'+:1+'%' and t.f=:2", wantValues: []any{"1", "1"}},
		{name: "b-2-6.", tpl: "|{like %t.f%} &{t.f}", input: map[string]any{"f": "1"}, wantSql: "or t.f like '%'+:1+'%' and t.f=:2", wantValues: []any{"1", "1"}},

		{name: "c-1-1.", tpl: "&{in f}", input: map[string]any{"f": []string{"1"}}, wantSql: "and f in ('1')", wantValues: nil},
		{name: "c-1-2.", tpl: "&{in t.f}", input: map[string]any{"f": []string{"1", "2"}}, wantSql: "and t.f in ('1','2')", wantValues: nil},
		{name: "c-1-3.", tpl: "&{in t.f}", input: map[string]any{"f": []string{"1", "'"}}, wantSql: "and t.f in ('1','''')", wantValues: nil},
		{name: "c-1-4.", tpl: "&{in t.f}", input: map[string]any{"f": []string{"1", "2", "'--"}}, wantSql: "and t.f in ('1','2','''--')", wantValues: nil},
		{name: "c-1-5.", tpl: "&{in f}", input: map[string]any{"f": []int{1, 2}}, wantSql: "and f in (1,2)", wantValues: nil},
		{name: "c-1-6.", tpl: "&{in t.f}", input: map[string]any{"f": []int{1, 2}}, wantSql: "and t.f in (1,2)", wantValues: nil},

		{name: "c-2-1.", tpl: "|{filed in f}", input: map[string]any{"f": []string{"1"}}, wantSql: "or filed in ('1')", wantValues: nil},
		{name: "c-2-2.", tpl: "|{t.field in f}", input: map[string]any{"f": []string{"1"}}, wantSql: "or t.field in ('1')", wantValues: nil},
		{name: "c-2-3.", tpl: "|{in f}", input: map[string]any{"f": []int{1, 2}}, wantSql: "or f in (1,2)", wantValues: nil},
		{name: "c-2-4.", tpl: "|{in t.f}", input: map[string]any{"f": []int{1, 2}}, wantSql: "or t.f in (1,2)", wantValues: nil},

		{name: "d-1-1.", tpl: "&{> f}", input: map[string]any{"f": "1"}, wantSql: "and f>:1", wantValues: []any{"1"}},
		{name: "d-1-2.", tpl: "&{>= t.f}", input: map[string]any{"f": "1"}, wantSql: "and t.f>=:1", wantValues: []any{"1"}},
		{name: "d-1-3.", tpl: "&{< f}", input: map[string]any{"f": "1"}, wantSql: "and f<:1", wantValues: []any{"1"}},
		{name: "d-1-4.", tpl: "&{<= t.f}", input: map[string]any{"f": "1"}, wantSql: "and t.f<=:1", wantValues: []any{"1"}},
		{name: "d-1-5.", tpl: "&{= t.f}", input: map[string]any{"f": "1"}, wantSql: "and t.f=:1", wantValues: []any{"1"}},

		{name: "d-2-1.", tpl: "&{field > f}", input: map[string]any{"f": "1"}, wantSql: "and field>:1", wantValues: []any{"1"}},
		{name: "d-2-2.", tpl: "&{t.field >= f}", input: map[string]any{"f": "1"}, wantSql: "and t.field>=:1", wantValues: []any{"1"}},
		{name: "d-2-3.", tpl: "&{field < f}", input: map[string]any{"f": "1"}, wantSql: "and field<:1", wantValues: []any{"1"}},
		{name: "d-2-4.", tpl: "&{t.field <= f}", input: map[string]any{"f": "1"}, wantSql: "and t.field<=:1", wantValues: []any{"1"}},
		{name: "d-2-5.", tpl: "&{t.field = f}", input: map[string]any{"f": "1"}, wantSql: "and t.field=:1", wantValues: []any{"1"}},
	}

	for _, tt := range tests {
		tt.template = fixedtemplate
		t.Run(tt.name, func(t *testing.T) {
			gotSql, gotValues, err := AnalyzeTPLFromCache(tt.template, tt.tpl, tt.input)
			if tt.wantErr != (err != nil) {
				t.Errorf("AnalyzeTPLFromCache() wantErr = %v, want %v", err, tt.wantErr)
			}
			if gotSql != tt.wantSql {
				t.Errorf("AnalyzeTPLFromCache() gotSql = %v, want %v", gotSql, tt.wantSql)
			}
			if !reflect.DeepEqual(gotValues, tt.wantValues) {
				t.Errorf("AnalyzeTPLFromCache() gotValues = %v, want %v", gotValues, tt.wantValues)
			}
		})
	}

}
