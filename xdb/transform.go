package xdb

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/zhiyunliu/golibs/xreflect"
)

// Translate 模板转换字符串
func Translate(tpl string, data map[string]any) string {
	return xTranslate(tpl, data)
}

func TranslateMap(tpl string, data map[string]string) string {
	return xTranslate(tpl, data)
}

// TranslateCallback 使用回调函数进行模板转换
// eg: str="aaa:@bbb:@{ccc}" callback=function(key) { return strings.ToUpper(key) } ==> "aaa:@bbb:CCC"
// @{aaa}, {aaa} 可以适配转换
func TranslateCallback(tpl string, callback func(param string) string) string {
	return translateWithCallback(tpl, callback)
}

// GenericTranslate 泛型模板转换字符串
func GenericTranslate[T any](tpl string, data T) string {
	var tmpData any = data

	if v, ok := tmpData.(map[string]any); ok {
		return xTranslate(tpl, v)
	}

	if v, ok := tmpData.(map[string]string); ok {
		return xTranslate(tpl, v)
	}

	mapVal, err := xreflect.AnyToMap(data, xreflect.WithMaxDepth(1))
	if err != nil {
		return ""
	}
	return xTranslate(tpl, mapVal)
}

// toString 高性能类型转换（覆盖常见类型）
func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(x).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(x).Uint(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(x).Float(), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprint(x) // 兜底处理
	}
}

func xTranslate[T any](template string, data map[string]T) string {
	return translateWithCallback(template, func(key string) string {
		if value, ok := data[key]; ok {
			return toString(value)
		}
		return ""
	})
}

// translateWithCallback 使用回调函数处理模板的核心逻辑
func translateWithCallback(template string, callback func(param string) string) string {
	var builder strings.Builder
	var keyBuilder strings.Builder
	inVar := false
	inBrace := false
	inSimpleBrace := false // 标记是否处于简单大括号模式

	for i, cnt := 0, len(template); i < cnt; i++ {
		c := template[i]

		switch {
		case c == '@' && !inVar:
			// 开始变量，只支持@{xxx}格式
			if i+1 < cnt && template[i+1] == '{' {
				inVar = true
				inBrace = true
				i++ // 跳过'{'
			} else {
				// 如果是单独的@符号而不是@{xxx}格式，则直接写入builder
				builder.WriteByte(c)
			}

		case c == '{' && !inVar:
			// 开始简单大括号模式 {variable}
			inVar = true
			inSimpleBrace = true

		case inVar && inBrace && c == '}':
			// 结束@{var}模式
			key := keyBuilder.String()
			value := callback(key)
			builder.WriteString(value)
			keyBuilder.Reset()
			inVar = false
			inBrace = false

		case inVar && inSimpleBrace && c == '}':
			// 结束{var}模式
			key := keyBuilder.String()
			value := callback(key)
			builder.WriteString(value)
			keyBuilder.Reset()
			inVar = false
			inSimpleBrace = false

		case inVar:
			// 收集变量名
			keyBuilder.WriteByte(c)

		default:
			// 普通字符
			builder.WriteByte(c)
		}
	}

	return builder.String()
}
