package expression

import "github.com/zhiyunliu/glue/xdb"

func initOperator() {
	xdb.NewOperator = NewDefaultOperator
}

type DefaultOperator struct {
	OperatorName           string
	ExpressionCallback     xdb.ExpressionCallback
	NormalizeValueCallback xdb.NormalizeValueCallback
}

func NewDefaultOperator(name string, callback xdb.ExpressionCallback, normalize xdb.NormalizeValueCallback) xdb.Operator {
	return &DefaultOperator{
		OperatorName:           name,
		ExpressionCallback:     callback,
		NormalizeValueCallback: normalize,
	}
}

func (d *DefaultOperator) Name() string {
	return d.OperatorName
}

func (d *DefaultOperator) Callback(valuer xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
	if d.ExpressionCallback == nil {
		return ""
	}
	return d.ExpressionCallback(valuer, param, phName, value)
}

func (d *DefaultOperator) NormalizeValue(valuer xdb.ExprName, param xdb.DBParam, value any) (newVal any, err xdb.MissError) {
	if d.NormalizeValueCallback == nil {
		return value, nil
	}
	return d.NormalizeValueCallback(valuer, param, value)
}
