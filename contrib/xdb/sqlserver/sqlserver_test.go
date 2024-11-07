package sqlserver

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/zhiyunliu/glue/xdb"
)

func TestSqlserverGetSQLContext(t *testing.T) {

	template, err := xdb.GetTemplate(Proto)
	if err != nil {
		return
	}
	//--------------------------------------------------------
	execSql := `select 1 from t where p1=@{p1}`
	execParam := map[string]any{"p1": 1}

	//1. check first
	query, execArgs, err := template.GetSQLContext(execSql, execParam)
	checkcase1(t, query, execArgs)

	//1 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase1(t, query, execArgs)

	//-------------------------------------------------------------

	execSql = `select 1 from t where p1=@{p1} and p2=@{p2}`
	execParam = map[string]any{"p1": 1, "p2": 2}

	//2. check first
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase2(t, query, execArgs)

	//2 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase2(t, query, execArgs)

	//-------------------------------------------------------------
	execSql = `select 1 from t where p1=@{p1} &{p2}`
	execParam = map[string]any{"p1": 1}

	//3. check and
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase3(t, execParam, query, execArgs)

	//2 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase3(t, execParam, query, execArgs)

	execParam = map[string]any{"p1": 1, "p2": 2}

	//3. check and
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase3(t, execParam, query, execArgs)

	//2 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase3(t, execParam, query, execArgs)

	execParam = map[string]any{"p1": 1}

	//3. check and
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase3(t, execParam, query, execArgs)

	execParam = map[string]any{"p1": 1, "p2": 2}

	//2 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase3(t, execParam, query, execArgs)

	// -------------------------------------------------------------
	execSql = `select 1 from t where p1=@{p1} |{p2}`
	execParam = map[string]any{"p1": 1}

	// 3. check and
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase4(t, execParam, query, execArgs)

	// 2 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase4(t, execParam, query, execArgs)

	execParam = map[string]any{"p1": 1, "p2": 2}

	// 3. check and
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase4(t, execParam, query, execArgs)

	// 2 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase4(t, execParam, query, execArgs)

	execParam = map[string]any{"p1": 1}

	// 3. check and
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase4(t, execParam, query, execArgs)

	execParam = map[string]any{"p1": 1, "p2": 2}

	// 2 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase4(t, execParam, query, execArgs)

	// -------------------------------------------------------------
	execSql = `select 1 from t where p1=@{p1} in (${p3})`
	execParam = map[string]any{"p1": 1, "p3": []int{1, 2}}

	// 3. check and
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase5(t, execParam, query, execArgs)

	// 2 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase5(t, execParam, query, execArgs)

	execParam = map[string]any{"p1": 1, "p3": []string{"a", "b", "c"}}

	// 3. check and
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase5(t, execParam, query, execArgs)

	// 2 check from cache
	query, execArgs, err = template.GetSQLContext(execSql, execParam)
	checkcase5(t, execParam, query, execArgs)

}

func checkcase1(t *testing.T, query string, execArgs []any) {
	if query != "select 1 from t where p1=@p_p1" {
		t.Error("1.1", query)
	}
	if len(execArgs) != 1 {
		t.Error("1.2", execArgs)
	}

	for i := range execArgs {
		if arg, ok := execArgs[i].(sql.NamedArg); !ok {
			t.Error("1.3", execArgs[i])
		} else {
			if arg.Name != "p_p1" {
				t.Error("1.4", arg.Name)
			}
			if arg.Value != 1 {
				t.Error("1.5", arg.Value)
			}
		}
	}
}

func checkcase2(t *testing.T, query string, execArgs []any) {
	if query != "select 1 from t where p1=@p_p1 and p2=@p_p2" {
		t.Error(".21", query)
	}

	if len(execArgs) != 2 {
		t.Error("2.2", execArgs)
	}

	for i := range execArgs {
		if arg, ok := execArgs[i].(sql.NamedArg); !ok {
			t.Error("2.3", execArgs[i])
		} else {
			if i == 0 {
				if arg.Name != "p_p1" {
					t.Error("2.4", arg.Name)
				}
				if arg.Value != 1 {
					t.Error("2.5", arg.Value)
				}
			}
			if i == 1 {
				if arg.Name != "p_p2" {
					t.Error("2.4", arg.Name)
				}
				if arg.Value != 2 {
					t.Error("2.5", arg.Value)
				}
			}
		}
	}
}

func checkcase3(t *testing.T, param map[string]any, query string, execArgs []any) {
	paramLen := len(param)
	query = strings.TrimSpace(query)

	rquery := "select 1 from t where p1=@p_p1 and p2=@p_p2"
	if paramLen == 1 {
		rquery = "select 1 from t where p1=@p_p1"
	}

	if query != rquery {
		t.Error("3.1", query)
	}

	if len(execArgs) != paramLen {
		t.Error("3.2", execArgs)
	}

	for i := range execArgs {
		if arg, ok := execArgs[i].(sql.NamedArg); !ok {
			t.Error("3.3", execArgs[i])
		} else {

			argName := strings.TrimPrefix(arg.Name, ArgumentPrefix)
			if param[argName] != arg.Value {
				t.Error("3.4", arg.Name, arg.Value)
			}
		}
	}
}

func checkcase4(t *testing.T, param map[string]any, query string, execArgs []any) {
	paramLen := len(param)
	query = strings.TrimSpace(query)

	rquery := "select 1 from t where p1=@p_p1 or p2=@p_p2"
	if paramLen == 1 {
		rquery = "select 1 from t where p1=@p_p1"
	}

	if query != rquery {
		t.Error("4.1", query)
	}

	if len(execArgs) != paramLen {
		t.Error("4.2", execArgs)
	}

	for i := range execArgs {
		if arg, ok := execArgs[i].(sql.NamedArg); !ok {
			t.Error("4.3", execArgs[i])
		} else {

			argName := strings.TrimPrefix(arg.Name, ArgumentPrefix)
			if param[argName] != arg.Value {
				t.Error("4.4", arg.Name, arg.Value)
			}
		}
	}
}

func checkcase5(t *testing.T, param map[string]any, query string, execArgs []any) {

	if len(execArgs) != 1 {
		t.Log("5.0", execArgs)
	}

	rquery := ""
	p3v := param["p3"]

	switch p3v.(type) {
	case []int:
		rquery = `select 1 from t where p1=@p_p1 in (1,2)`
	case []string:
		rquery = `select 1 from t where p1=@p_p1 in ('a','b','c')`

	}

	if query != rquery {
		t.Error("5.1", query)
	}

	if len(execArgs) != 1 {
		t.Error("5.2", execArgs)
	}

	for i := range execArgs {
		if arg, ok := execArgs[i].(sql.NamedArg); !ok {
			t.Error("5.3", execArgs[i])
		} else {

			argName := strings.TrimPrefix(arg.Name, ArgumentPrefix)
			if param[argName] != arg.Value {
				t.Error("5.4", arg.Name, arg.Value)
			}
		}
	}
}
