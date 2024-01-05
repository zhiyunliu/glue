package sqlserver

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
)

func TestMssqlContext_AnalyzeTPL(t *testing.T) {
	type args struct {
		template string
		input    map[string]interface{}
		ph       tpl.Placeholder
	}

	var tint int = 1
	var ctx *MssqlContext = &MssqlContext{
		name:    "mssql",
		prefix:  "p_",
		symbols: newMssqlSymbols(tpl.DefaultOperator.Clone()),
	}
	var ph tpl.Placeholder = ctx.Placeholder()
	tests := []struct {
		name  string
		args  args
		want  string
		want1 *tpl.ReplaceItem
	}{
		{name: "0.", args: args{template: `select * from a where 1=1 &{a.out_city_no}`, input: map[string]interface{}{}, ph: ph},
			want: "select * from a where 1=1 ",
			want1: &tpl.ReplaceItem{
				NameCache:   map[string]string{},
				Placeholder: ph,
				HasAndOper:  true,
			}},
		{name: "1.", args: args{template: `select * from a where 1=1 &{a.out_city_no}`, input: map[string]interface{}{"out_city_no": 1}, ph: ph},
			want: "select * from a where 1=1 and a.out_city_no=@p_out_city_no",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: 1}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  true,
			}},

		{name: "2.", args: args{template: `select * from a where 1=1 &{a.out_city_no}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "select * from a where 1=1 and a.out_city_no=@p_out_city_no",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  true,
			}},

		{name: "3.", args: args{template: `select * from a where 1=1 |{a.out_city_no}`, input: map[string]interface{}{"out_city_no": 1}, ph: ph},
			want: "select * from a where 1=1 or a.out_city_no=@p_out_city_no",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: 1}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasOrOper:   true,
			}},

		{name: "4.", args: args{template: `select * from a where 1=1 |{a.out_city_no}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "select * from a where 1=1 or a.out_city_no=@p_out_city_no",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasOrOper:   true,
			}},

		{name: "5.", args: args{template: `select * from a where 1=1 |{a.out_city_no} &{a.out_city_no}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "select * from a where 1=1 or a.out_city_no=@p_out_city_no and a.out_city_no=@p_out_city_no",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  true,
				HasOrOper:   true,
			}},

		{name: "1a.", args: args{template: `select * from a where 1=1 |{like a.out_city_no}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "select * from a where 1=1 or a.out_city_no like @p_out_city_no",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  false,
				HasOrOper:   true,
			}},
		{name: "1b.", args: args{template: `select * from a where 1=1 &{like a.out_city_no}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "select * from a where 1=1 and a.out_city_no like @p_out_city_no",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  true,
				HasOrOper:   false,
			}},

		{name: "1c.", args: args{template: `select * from a where 1=1 |{like %a.out_city_no}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "select * from a where 1=1 or a.out_city_no like '%'+@p_out_city_no",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  false,
				HasOrOper:   true,
			}},
		{name: "1d.", args: args{template: `select * from a where 1=1 &{like %a.out_city_no}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "select * from a where 1=1 and a.out_city_no like '%'+@p_out_city_no",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  true,
				HasOrOper:   false,
			}},

		{name: "1e.", args: args{template: `select * from a where 1=1 |{like %a.out_city_no%}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "select * from a where 1=1 or a.out_city_no like '%'+@p_out_city_no+'%'",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  false,
				HasOrOper:   true,
			}},
		{name: "1f.", args: args{template: `select * from a where 1=1 &{like %a.out_city_no%}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "select * from a where 1=1 and a.out_city_no like '%'+@p_out_city_no+'%'",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  true,
				HasOrOper:   false,
			}},

		{name: "2f.", args: args{template: `&{like %a.out_city_no%} |{like %a.out_city_no%} &{like %a.out_city_no%}`, input: map[string]interface{}{"out_city_no": &tint}, ph: ph},
			want: "and a.out_city_no like '%'+@p_out_city_no+'%' or a.out_city_no like '%'+@p_out_city_no+'%' and a.out_city_no like '%'+@p_out_city_no+'%'",
			want1: &tpl.ReplaceItem{Names: []string{"out_city_no"},
				Values:      []any{sql.NamedArg{Name: "p_out_city_no", Value: &tint}},
				NameCache:   map[string]string{"out_city_no": "@p_out_city_no"},
				Placeholder: ph,
				HasAndOper:  true,
				HasOrOper:   true,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, _ := ctx.AnalyzeTPL(tt.args.template, tt.args.input, tt.args.ph)
			if got != tt.want {
				t.Errorf("MssqlContext.AnalyzeTPL() got[%v], want[%v]", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MssqlContext.AnalyzeTPL() got1[%v], want1[%v]", got1, tt.want1)
			}
		})
	}
}
