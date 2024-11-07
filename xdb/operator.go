package xdb

import (
	"sync"
)

type OperatorCallback func(valuer ExpressionValuer, param DBParam, argName string, value any) string

type OperatorMap interface {
	Store(name string, callback OperatorCallback)
	Load(name string) (OperatorCallback, bool)
	Clone() OperatorMap
}

type operatorMap struct {
	syncMap *sync.Map
}

func NewOperatorMap() OperatorMap {
	return &operatorMap{
		syncMap: &sync.Map{},
	}
}

func (m *operatorMap) Store(name string, callback OperatorCallback) {
	m.syncMap.Store(name, callback)
}

func (m *operatorMap) Load(name string) (OperatorCallback, bool) {
	callback, ok := m.syncMap.Load(name)

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
