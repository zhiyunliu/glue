package xdb

import (
	"context"
	"fmt"

	"github.com/zhiyunliu/glue/config"
)

type TransactionCallback func(dbObj Executer) error

// IDB 数据库操作接口
type IDB interface {
	Executer
	Begin() (ITrans, error)
	Close() error
	GetImpl() interface{}
	Transaction(TransactionCallback) error
}

// ITrans 数据库事务接口
type ITrans interface {
	Executer
	Rollback() error
	Commit() error
}

// Executer 数据库操作对象集合
type Executer interface {
	Query(ctx context.Context, sql string, input any, opts ...TemplateOption) (data Rows, err error)
	Multi(ctx context.Context, sql string, input any, opts ...TemplateOption) (data []Rows, err error)
	First(ctx context.Context, sql string, input any, opts ...TemplateOption) (data Row, err error)
	Scalar(ctx context.Context, sql string, input any, opts ...TemplateOption) (data interface{}, err error)
	Exec(ctx context.Context, sql string, input any, opts ...TemplateOption) (r Result, err error)

	QueryAs(ctx context.Context, sql string, input any, result any, opts ...TemplateOption) (err error)
	FirstAs(ctx context.Context, sql string, input any, result any, opts ...TemplateOption) (err error)
}

// dbResover 定义配置文件转换方法
type Resolver interface {
	Name() string
	Resolve(connName string, setting config.Config, opts ...Option) (interface{}, error)
}

var dbResolvers = make(map[string]Resolver)

// Register 注册配置文件适配器
func Register(resolver Resolver) {
	proto := resolver.Name()
	if _, ok := dbResolvers[proto]; ok {
		panic(fmt.Errorf("db: 不能重复注册:%s", proto))
	}
	dbResolvers[proto] = resolver
}

// Deregister 清理配置适配器
func Deregister(name string) {
	delete(dbResolvers, name)
}

// newDB 根据适配器名称及参数返回配置处理器
func newDB(connName string, setting config.Config, opts ...Option) (interface{}, error) {
	val := setting.Value("proto")
	proto := val.String()
	resolver, ok := dbResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("db: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(connName, setting, opts...)
}
