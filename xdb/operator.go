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

// OperatorMap 操作符映射接口
type OperatorMap interface {
	Store(name string, callback OperatorCallback)
	Load(name string) (OperatorCallback, bool)
	Clone() OperatorMap
	Range(func(name string, callback OperatorCallback) bool)
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
		operMap.Store(oper.Name(), oper.Callback)
	}
	return operMap
}

func (m *operatorMap) Store(name string, callback OperatorCallback) {
	m.syncMap.Store(name, callback)
}

func (m *operatorMap) Load(name string) (OperatorCallback, bool) {
	callback, ok := m.syncMap.Load(name)
	if !ok {
		return nil, ok
	}

	return callback.(OperatorCallback), ok
}

func (m *operatorMap) Clone() OperatorMap {
	clone := NewOperatorMap()
	m.syncMap.Range(func(key, value any) bool {
		clone.Store(key.(string), value.(OperatorCallback))
		return true
	})
	return clone
}

func (m *operatorMap) Range(f func(name string, callback OperatorCallback) bool) {
	m.syncMap.Range(func(key, value any) bool {
		return f(key.(string), value.(OperatorCallback))
	})
}
