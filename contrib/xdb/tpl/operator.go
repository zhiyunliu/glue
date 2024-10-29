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
	DefaultOperator.Store("=", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s=%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.Store(">", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s>%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.Store(">=", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s>=%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.Store("<", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s<%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.Store("<=", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s<=%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.Store("like", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s like %s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.Store("%like", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s like '%%'+%s", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.Store("like%", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s like %s+'%%'", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.Store("%like%", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s like '%%'+%s+'%%'", getConcat(symbol), fullkey, argName)
	})

	DefaultOperator.Store("in", func(symbol, fullkey, argName string) string {
		return fmt.Sprintf("%s %s in (select value from string_split(%s,','))", getConcat(symbol), fullkey, argName) //有注入风险
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

// func (m *operatorMap) Register(oper Operator) error {
// 	_, loaded := m.syncMap.LoadOrStore(oper.Name(), oper.Callback)
// 	if loaded {
// 		return nil
// 	}

// 	return nil
// }

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
