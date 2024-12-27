package sqlserver

import (
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xreflect"
)

var (
	DefaultDbTypeHandler = []xdb.StmtDbTypeHandler{
		&varcharHandler{},
		&varcharMaxHandler{},
		&nvarcharMaxHandler{},
		&tvpHandler{},
	}
)

type varcharHandler struct {
}

func (h *varcharHandler) Name() string {
	return "varchar"
}
func (h *varcharHandler) Handle(param any, args []string) any {
	tmpval := xreflect.GetString(param)
	return mssql.VarChar(tmpval)
}

type varcharMaxHandler struct {
}

func (h *varcharMaxHandler) Name() string {
	return "varcharmax"
}
func (h *varcharMaxHandler) Handle(param any, args []string) any {
	tmpval := xreflect.GetString(param)
	return mssql.VarCharMax(tmpval)
}

type nvarcharMaxHandler struct {
}

func (h *nvarcharMaxHandler) Name() string {
	return "nvarcharmax"
}
func (h *nvarcharMaxHandler) Handle(param any, args []string) any {
	tmpval := xreflect.GetString(param)
	return mssql.NVarCharMax(tmpval)
}

type tvpHandler struct {
}

func (h *tvpHandler) Name() string {
	return "tvp"
}
func (h *tvpHandler) Handle(param any, args []string) any {
	//args = tvp=typename
	//args = ["tvp","typename"]
	if len(args) != 2 {
		return param
	}
	return mssql.TVP{
		TypeName: args[1],
		Value:    param,
	}
}
