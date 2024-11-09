package xdb

import (
	"sync"
)

// OperatorCallback 操作符回调函数
type OperatorCallback func(valuer ExpressionValuer, param DBParam, phName string, value any) string

// Operator 操作符处理接口
type Operator interface {
	Name() string
	Callback(valuer ExpressionValuer, param DBParam, phName string, value any) string
}

type DefaultOperator struct {
	name     string
	callback OperatorCallback
}

func NewDefaultOperator(name string, callback OperatorCallback) Operator {
	return &DefaultOperator{name: name, callback: callback}
}

func (d *DefaultOperator) Name() string {
	return ""
}

func (d *DefaultOperator) Callback(valuer ExpressionValuer, param DBParam, phName string, value any) string {
	return ""
}

// OperatorMap 操作符映射接口
type OperatorMap interface {
	//Store(name string, callback OperatorCallback)
	Load(name string) (Operator, bool)
	Clone() OperatorMap
	Range(func(name string, callback Operator) bool)
}

type operatorMap struct {
	syncMap *sync.Map
}

// NewOperatorMap 创建操作符映射
func NewOperatorMap(operators ...Operator) OperatorMap {
	operMap := &operatorMap{
		syncMap: &sync.Map{},
	}
	for _, oper := range operators {
		operMap.syncMap.Store(oper.Name(), oper)
	}
	return operMap
}

func (m *operatorMap) Load(name string) (Operator, bool) {
	callback, ok := m.syncMap.Load(name)
	if !ok {
		return nil, ok
	}

	return callback.(Operator), ok
}

func (m *operatorMap) Clone() OperatorMap {
	clone := &operatorMap{
		syncMap: &sync.Map{},
	}
	m.syncMap.Range(func(key, value any) bool {
		clone.syncMap.Store(key.(string), value.(OperatorCallback))
		return true
	})
	return clone
}

func (m *operatorMap) Range(f func(name string, operator Operator) bool) {
	m.syncMap.Range(func(key, value any) bool {
		return f(key.(string), value.(Operator))
	})
}
