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

type symbolsMap struct {
	pattern string
	syncMap *sync.Map
}

func NewSymbolMap() SymbolMap {
	return &symbolsMap{
		pattern: TotalPattern,
		syncMap: &sync.Map{},
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

func (m *symbolsMap) Store(name string, callback SymbolCallback) {
	m.syncMap.Store(name, callback)
}

func (m *symbolsMap) LoadOrStore(name string, callback SymbolCallback) (loaded bool) {
	_, loaded = m.syncMap.LoadOrStore(name, callback)
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
	clone := NewSymbolMap()
	m.syncMap.Range(func(key, value any) bool {
		clone.Store(key.(string), value.(SymbolCallback))
		return true
	})
	return clone
}

var defaultSymbols SymbolMap //  Symbols

func init() {
	defaultSymbols = NewSymbolMap()
	defaultSymbols.Store("@", func(input DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissParamError) {
		propName := GetPropName(fullKey)
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

	defaultSymbols.Store("&", func(input DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissParamError) {
		propName := GetPropName(fullKey)
		argName, value, err := input.Get(propName, item.Placeholder)
		if err != nil {
			return "", err
		}
		item.HasAndOper = true
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf(" and %s=%s", fullKey, argName), nil
		}
		return "", nil
	})

	defaultSymbols.Store("|", func(input DBParam, fullKey string, item *ReplaceItem) (string, xdb.MissParamError) {
		propName := GetPropName(fullKey)
		argName, value, err := input.Get(propName, item.Placeholder)
		if err != nil {
			return "", err
		}
		item.HasOrOper = true
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf(" or %s=%s", fullKey, argName), nil
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
