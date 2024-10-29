package tpl

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/contrib/xdb/prop"
	"github.com/zhiyunliu/glue/xdb"
)

// 根据表达式获取
var GetPropMatchValuer func(fullKey string, opts *xdb.PropOptions) (matcher xdb.PropValuer)

func init() {
	GetPropMatchValuer = prop.DefaultGetPropMatchValuer
}

type symbolsMap struct {
	patternList []string
	symbolMap   *sync.Map
	extMap      *sync.Map
	operMap     OperatorMap
}

func NewSymbolMap(operMap OperatorMap) SymbolMap {
	return &symbolsMap{
		patternList: DefaultPatternList,
		symbolMap:   &sync.Map{},
		extMap:      &sync.Map{},
		operMap:     operMap,
	}
}

func (m *symbolsMap) GetPattern() string {

	list := make([]string, 0, len(m.patternList))

	list = append(list, m.patternList...)

	m.extMap.Range(func(key, value any) bool {
		list = append(list, value.(Symbol).GetPattern())
		return true
	})

	return strings.Join(list, "|")
}

func (m *symbolsMap) RegisterSymbol(symbol Symbol) error {
	m.StoreSymbol(symbol.Name(), symbol.Callback)

	pattern := strings.TrimSpace(symbol.GetPattern())
	if pattern != "" {
		_, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("表达式:%s,不是有效的正则,%w", pattern, err)
		}
		m.extMap.Store(symbol.Name(), symbol)
	}
	return nil
}

func (m *symbolsMap) RegisterOperator(oper Operator) error {
	if oper == nil {
		return nil
	}
	m.operMap.Store(oper.Name(), oper.Callback)
	return nil
}

func (m *symbolsMap) StoreSymbol(name string, callback SymbolCallback) {
	m.symbolMap.Store(name, callback)
}

func (m *symbolsMap) LoadSymbol(name string) (SymbolCallback, bool) {
	callback, ok := m.symbolMap.Load(name)
	return callback.(SymbolCallback), ok
}
func (m *symbolsMap) Delete(name string) {
	m.symbolMap.Delete(name)
}

func (m *symbolsMap) Clone() SymbolMap {
	clone := NewSymbolMap(m.operMap.Clone())

	m.extMap.Range(func(key, value any) bool {
		clone.RegisterSymbol(value.(Symbol))
		return true
	})

	m.symbolMap.Range(func(key, value any) bool {
		clone.StoreSymbol(key.(string), value.(SymbolCallback))
		return true
	})

	return clone
}

var defaultSymbols SymbolMap //  Symbols

func init() {
	defaultSymbols = NewSymbolMap(DefaultOperator)
	defaultSymbols.StoreSymbol(SymbolAt, func(symbolMap SymbolMap, input xdb.DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissError) {
		matcher := GetPropMatchValuer(fullKey, item.PropOpts)
		if matcher == nil {
			return "", xdb.NewMissPropError(fullKey)
		}
		propName := matcher.GetPropName()

		argName, value, err := input.Get(propName, item.Placeholder)
		if err != nil {
			return "", err
		}
		if !xdb.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
		} else {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, nil)
		}
		return argName, nil
	})

	defaultSymbols.StoreSymbol(SymbolAnd, func(symbolMap SymbolMap, input xdb.DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissError) {
		item.HasAndOper = true

		matcher := GetPropMatchValuer(fullKey, item.PropOpts)
		if matcher == nil {
			return "", xdb.NewMissPropError(fullKey)
		}
		propName := matcher.GetPropName()

		argName, value, _ := input.Get(propName, item.Placeholder)
		if !xdb.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return matcher.Build(SymbolAnd, input, argName)
		}
		return "", nil
	})

	defaultSymbols.StoreSymbol(SymbolOr, func(symbolMap SymbolMap, input xdb.DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissError) {
		item.HasOrOper = true

		matcher := GetPropMatchValuer(fullKey, item.PropOpts)
		if matcher == nil {
			return "", xdb.NewMissPropError(fullKey)
		}
		propName := matcher.GetPropName()

		argName, value, _ := input.Get(propName, item.Placeholder)

		if !xdb.IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return matcher.Build(SymbolOr, input, argName)
			//return fmt.Sprintf(" or %s=%s", fullKey, argName), nil
		}
		return "", nil
	})
}
