package xdb

import (
	"fmt"
	"strings"
	"sync"
)

var (
	tpls sync.Map
	// 新建一个模板匹配器
	NewTemplateMatcher func(matchers ...ExpressionMatcher) TemplateMatcher
)

// Template 模板上下文
type SQLTemplate interface {
	Name() string
	Placeholder() Placeholder
	//获取sql
	GetSQLContext(tpl string, input map[string]any, opts ...TemplateOption) (sql string, args []any, err error)
	//注册表达式匹配解析器
	RegistExpressionMatcher(matchers ...ExpressionMatcher)
	//处理一般表达式
	HandleExpr(item SqlState, sqlTpl string, param DBParam) (sql string, err error)
	//获取sql状态
	GetSqlState(*TemplateOptions) SqlState
	//sql状态释放
	ReleaseSqlState(SqlState)
}

type ExpressionCache interface {
	Build(SqlState, DBParam) (sql string, err error)
}

// RegistTemplate 注册模板
func RegistTemplate(tpl SQLTemplate) (err error) {
	name := strings.ToLower(tpl.Name())
	tpls.Store(name, tpl)
	return
}

// GetDBTemplate 获取数据库上下文操作
func GetTemplate(name string) (SQLTemplate, error) {
	if v, ok := tpls.Load(strings.ToLower(name)); ok {
		return v.(SQLTemplate), nil
	}
	return nil, fmt.Errorf("不支持的数据库类型:%s", name)
}

// RegistExpressionMatcher 给数据库注册语法解析
func RegistExpressionMatcher(proto string, matcher ExpressionMatcher) (err error) {
	tmpl, err := GetTemplate(proto)
	if err != nil {
		return err
	}
	tmpl.RegistExpressionMatcher(matcher)
	return
}

// 表达式解析选项
type TemplateOptions struct {
	UseExprCache bool
}

type TemplateOption func(*TemplateOptions)

// 使用解析缓存
func WithExprCache(use bool) TemplateOption {
	return func(o *TemplateOptions) {
		o.UseExprCache = use
	}
}

type MatcherOptions struct {
	BuildCallback ExpressionBuildCallback
	OperatorMap   OperatorMap
}
type MatcherOption func(*MatcherOptions)

// WithBuildCallback 制定matcher的表达式生成回调
func WithBuildCallback(callback ExpressionBuildCallback) MatcherOption {
	return func(mo *MatcherOptions) {
		mo.BuildCallback = callback
	}
}

// WithOperatorMap 制定matcher的符号处理函数 与WithOperator 二选一
func WithOperatorMap(operatorMap OperatorMap) MatcherOption {
	return func(mo *MatcherOptions) {
		mo.OperatorMap = operatorMap
	}
}

// WithOperator 增加一个符号处理函数 与WithOperatorMap 二选一
func WithOperator(operator ...Operator) MatcherOption {
	return func(mo *MatcherOptions) {
		mo.OperatorMap = NewOperatorMap(operator...)
	}
}

// TemplateMatcher 模板匹配器
type TemplateMatcher interface {
	// RegistMatcher 注册表达式匹配器
	RegistMatcher(matcher ...ExpressionMatcher)
	// GenerateSQL 根据模板生成SQL语句
	GenerateSQL(item SqlState, sqlTpl string, input DBParam) (sql string, err error)
}
