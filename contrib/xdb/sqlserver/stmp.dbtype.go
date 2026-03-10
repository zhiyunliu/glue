package sqlserver

import (
	"reflect"

	mssql "github.com/microsoft/go-mssqldb"
	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xreflect"
)

var (
	DefaultDbTypeHandler = []xdb.StmtDbTypeHandler{
		&varcharHandler{},
		&varcharMaxHandler{},
		&nvarcharMaxHandler{},
		&tvpHandler{},
		&contribxdb.StmtDbTypeOutputHandler{},
	}
)

type varcharHandler struct {
}

func (h *varcharHandler) Name() string {
	return "varchar"
}
func (h *varcharHandler) Handle(_ string, param any, _ reflect.Value, _ []string) (any, error) {
	tmpval := xreflect.GetString(param)
	return mssql.VarChar(tmpval), nil
}

type varcharMaxHandler struct {
}

func (h *varcharMaxHandler) Name() string {
	return "varcharmax"
}
func (h *varcharMaxHandler) Handle(_ string, param any, _ reflect.Value, _ []string) (any, error) {
	tmpval := xreflect.GetString(param)
	return mssql.VarCharMax(tmpval), nil
}

type nvarcharMaxHandler struct {
}

func (h *nvarcharMaxHandler) Name() string {
	return "nvarcharmax"
}
func (h *nvarcharMaxHandler) Handle(_ string, param any, _ reflect.Value, _ []string) (any, error) {
	tmpval := xreflect.GetString(param)
	return mssql.NVarCharMax(tmpval), nil
}

type tvpHandler struct {
}

func (h *tvpHandler) Name() string {
	return "tvp"
}
func (h *tvpHandler) Handle(_ string, param any, _ reflect.Value, args []string) (any, error) {
	//args = tvp=typename
	//args = ["tvp","typename"]
	if len(args) != 2 {
		return param, nil
	}
	return mssql.TVP{
		TypeName: args[1],
		Value:    param,
	}, nil
}
