package expression

import "github.com/zhiyunliu/glue/xdb"

func initOperator() {
	xdb.NewOperator = NewDefaultOperator
}

type DefaultOperator struct {
	name     string
	callback xdb.OperatorCallback
}

func NewDefaultOperator(name string, callback xdb.OperatorCallback) xdb.Operator {
	return &DefaultOperator{name: name, callback: callback}
}

func (d *DefaultOperator) Name() string {
	return d.name
}

func (d *DefaultOperator) Callback(valuer xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
	return d.callback(valuer, param, phName, value)
}
