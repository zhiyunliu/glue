package tpl

import (
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

func initStmtDbType() {
	xdb.NewStmtDbTypeProcessor = NewDefaultStmtDbTypeProcessor
}

type DefaultStmtDbTypeProcessor struct {
	exprCache *sync.Map
}

func NewDefaultStmtDbTypeProcessor(handlers ...xdb.StmtDbTypeHandler) xdb.StmtDbTypeProcessor {
	processor := &DefaultStmtDbTypeProcessor{
		exprCache: &sync.Map{},
	}
	processor.RegistHandler(handlers...)
	return processor
}

func (processor *DefaultStmtDbTypeProcessor) RegistHandler(handlers ...xdb.StmtDbTypeHandler) {
	for _, handler := range handlers {
		processor.exprCache.Store(handler.Name(), handler)
	}
}

func (processor *DefaultStmtDbTypeProcessor) Process(param any, tagOpts xdb.TagOptions) any {
	argsInfo, ok := tagOpts.GetArgsInfo("dbtype")
	if !ok {
		return param
	}
	dbtype := argsInfo[0]

	v, ok := processor.exprCache.Load(dbtype)
	if !ok || v == nil {
		return param
	}
	return v.(xdb.StmtDbTypeHandler).Handle(param, argsInfo)
}
