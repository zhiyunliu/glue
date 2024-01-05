package tpl

import (
	"fmt"
	"sync"
)

type operatorMap struct {
	syncMap *sync.Map
}

func NewOperatorMap() OperatorMap {
	return &operatorMap{
		syncMap: &sync.Map{},
	}
}

var DefaultOperator OperatorMap

func init() {
	DefaultOperator = NewOperatorMap()
	DefaultOperator.LoadOrStore("=", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s=%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.LoadOrStore(">", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s>%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.LoadOrStore(">=", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s>=%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.LoadOrStore("<", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s<%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.LoadOrStore("<=", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s<=%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.LoadOrStore("like", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s like %s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.LoadOrStore("%like", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s like '%%'+%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.LoadOrStore("like%", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s like %s+'%%'", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.LoadOrStore("%like%", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s like '%%'+%s+'%%'", getConcat(symbol), fullkey, argName)
	})

}

func getConcat(symbol string) (concat string) {
	switch symbol {
	case SymbolAnd:
		return "and"
	case SymbolOr:
		return "or"
	default:
		return ""
	}
}

func (m *operatorMap) Register(oper Operator) error {
	loaded := m.LoadOrStore(oper.Name(), oper.Callback)
	if loaded {
		return nil
	}

	return nil
}

func (m *operatorMap) LoadOrStore(name string, callback OperatorCallback) (loaded bool) {
	_, loaded = m.syncMap.LoadOrStore(name, callback)
	return
}

func (m *operatorMap) Load(name string) (OperatorCallback, bool) {
	callback, ok := m.syncMap.Load(name)
	return callback.(OperatorCallback), ok
}

func (m *operatorMap) Clone() OperatorMap {
	clone := NewOperatorMap()
	m.syncMap.Range(func(key, value any) bool {
		clone.LoadOrStore(key.(string), value.(OperatorCallback))
		return true
	})
	return clone
}
