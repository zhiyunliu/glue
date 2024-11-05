package xdb

import "sync"

// 符号回调函数
type SymbolCallback func(SymbolMap, DBParam, string, *SqlScene) (string, MissError)

type SymbolMap interface {
	Regist(Symbol)
	Load(name string) (Symbol, bool)
	Delete(name string)
	Clone() SymbolMap
}

type Symbol interface {
	Name() string
	Concat() string
	Callback(item *SqlScene, valuer ExpressionValuer, input DBParam) (string, MissError)
}

type symbolsMap struct {
	symbolMap *sync.Map
}

func NewSymbolMap(symbols ...Symbol) SymbolMap {
	var mapSymbols = &symbolsMap{
		symbolMap: &sync.Map{},
	}

	for i := range symbols {
		if symbols[i] == nil {
			continue
		}
		mapSymbols.Regist(symbols[i])
	}

	return mapSymbols
}

func (m *symbolsMap) Regist(symbol Symbol) {
	m.symbolMap.Store(symbol.Name(), symbol)
}

func (m *symbolsMap) Load(name string) (Symbol, bool) {
	callback, ok := m.symbolMap.Load(name)
	return callback.(Symbol), ok
}
func (m *symbolsMap) Delete(name string) {
	m.symbolMap.Delete(name)
}

func (m *symbolsMap) Clone() SymbolMap {
	clone := NewSymbolMap()
	m.symbolMap.Range(func(key, value any) bool {
		clone.Regist(value.(Symbol))
		return true
	})

	return clone
}
