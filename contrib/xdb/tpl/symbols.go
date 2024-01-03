package tpl

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

// 根据表达式获取
var GetPropName func(fullKey string) (field, propName, oper string)

func init() {
	GetPropName = DefaultGetPropName
}

type symbolsMap struct {
	pattern string
	syncMap *sync.Map
	operMap OperatorMap
}

func NewSymbolMap(operMap OperatorMap) SymbolMap {
	return &symbolsMap{
		pattern: TotalPattern,
		syncMap: &sync.Map{},
		operMap: operMap,
	}
}

func (m *symbolsMap) GetPattern() string {
	return m.pattern
}

func (m *symbolsMap) Register(symbol Symbol) error {
	loaded := m.LoadOrStore(symbol.Name(), symbol.Callback)
	if loaded {
		return nil
	}
	pattern := strings.TrimSpace(symbol.GetPattern())
	if pattern != "" {
		_, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("表达式:%s,不是有效的正则,%w", pattern, err)
		}
		m.pattern = m.pattern + "|" + pattern
	}
	return nil
}

func (m *symbolsMap) Operator(oper Operator) error {
	if oper == nil {
		return nil
	}
	m.operMap.LoadOrStore(oper.Name(), oper.Callback)
	return nil
}

func (m *symbolsMap) LoadOrStore(name string, callback SymbolCallback) (loaded bool) {
	_, loaded = m.syncMap.LoadOrStore(name, callback)
	return
}

func (m *symbolsMap) LoadOperator(oper string) (callback OperatorCallback, loaded bool) {
	callback, loaded = m.operMap.Load(oper)
	return
}
func (m *symbolsMap) Delete(name string) {
	m.syncMap.Delete(name)
}

func (m *symbolsMap) Load(name string) (SymbolCallback, bool) {
	callback, ok := m.syncMap.Load(name)
	return callback.(SymbolCallback), ok
}

func (m *symbolsMap) Clone() SymbolMap {
	clone := NewSymbolMap(m.operMap.Clone())
	m.syncMap.Range(func(key, value any) bool {
		clone.LoadOrStore(key.(string), value.(SymbolCallback))
		return true
	})
	return clone
}

var defaultSymbols SymbolMap //  Symbols

func init() {
	defaultSymbols = NewSymbolMap(DefaultOperator)
	defaultSymbols.LoadOrStore(SymbolAt, func(input DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissError) {
		_, propName, _ := GetPropName(fullKey)
		argName, value, err := input.Get(propName, item.Placeholder)
		if err != nil {
			return "", err
		}
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
		} else {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, nil)
		}
		return argName, nil
	})

	defaultSymbols.LoadOrStore(SymbolAnd, func(input DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissError) {
		item.HasAndOper = true

		fullField, propName, oper := GetPropName(fullKey)
		opercall, ok := defaultSymbols.LoadOperator(oper)
		if !ok {
			return "", xdb.NewMissOperError(oper)
		}

		argName, value, _ := input.Get(propName, item.Placeholder)
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return opercall(SymbolAnd, fullField, argName), nil
			//return fmt.Sprintf(" and %s=%s", fullKey, argName), nil
		}
		return "", nil
	})

	defaultSymbols.LoadOrStore(SymbolOr, func(input DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissError) {
		item.HasOrOper = true

		fullField, propName, oper := GetPropName(fullKey)
		opercall, ok := defaultSymbols.LoadOperator(oper)
		if !ok {
			return "", xdb.NewMissOperError(oper)
		}

		argName, value, _ := input.Get(propName, item.Placeholder)

		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return opercall(SymbolOr, fullField, argName), nil
			//return fmt.Sprintf(" or %s=%s", fullKey, argName), nil
		}
		return "", nil
	})
}

func IsNil(input interface{}) bool {
	if input == nil {
		return true
	}

	if arg, ok := input.(sql.NamedArg); ok {
		input = arg.Value
	}
	if arg, ok := input.(*sql.NamedArg); ok {
		input = arg.Value
	}
	if input == nil {
		return true
	}
	if fmt.Sprintf("%v", input) == "" {
		return true
	}
	rv := reflect.ValueOf(input)
	if rv.Kind() == reflect.Ptr {
		return rv.IsNil()
	}
	return false
}

// field, tbl.field , tbl.field like , tbl.field >=
func DefaultGetPropName(fullKey string) (fullField, propName, oper string) {
	propName = strings.TrimSpace(fullKey)
	fullField = propName
	idx := strings.Index(propName, " ")
	if idx < 0 {
		if strings.Index(propName, ".") > 0 {
			propName = strings.Split(propName, ".")[1]
		}
		oper = "="
		return fullField, propName, oper
	}

	parties := strings.Split(propName, " ")

	oper = parties[0]
	filed := parties[len(parties)-1]

	if strings.HasPrefix(filed, "%") {
		oper = "%" + oper
	}
	if strings.HasSuffix(filed, "%") {
		oper = oper + "%"
	}

	propName = strings.Trim(filed, "%")
	fullField = propName
	if strings.Index(propName, ".") > 0 {
		propName = strings.Split(propName, ".")[1]
	}
	return fullField, propName, oper
}
