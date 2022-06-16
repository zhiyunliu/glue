package xdb

import (
	"context"
	"fmt"

	"github.com/zhiyunliu/glue/config"
)

//IDB 数据库操作接口
type IDB interface {
	Executer
	Begin() (ITrans, error)
	Close() error
}

//ITrans 数据库事务接口
type ITrans interface {
	Executer
	Rollback() error
	Commit() error
}

//Executer 数据库操作对象集合
type Executer interface {
	Query(ctx context.Context, sql string, input map[string]interface{}) (data Rows, err error)
	Multi(ctx context.Context, sql string, input map[string]interface{}) (data []Rows, err error)
	First(ctx context.Context, sql string, input map[string]interface{}) (data Row, err error)
	Scalar(ctx context.Context, sql string, input map[string]interface{}) (data interface{}, err error)
	Exec(ctx context.Context, sql string, input map[string]interface{}) (r Result, err error)
	//StoredProc(procName string, input map[string]interface{}) (r Result, err error)
}

//dbResover 定义配置文件转换方法
type dbResover interface {
	Name() string
	Resolve(setting config.Config) (IDB, error)
}

var dbResolvers = make(map[string]dbResover)

//Register 注册配置文件适配器
func Register(resolver dbResover) {
	proto := resolver.Name()
	if _, ok := dbResolvers[proto]; ok {
		panic(fmt.Errorf("db: 不能重复注册:%s", proto))
	}
	dbResolvers[proto] = resolver
}

//Deregister 清理配置适配器
func Deregister(name string) {
	delete(dbResolvers, name)
}

//newDB 根据适配器名称及参数返回配置处理器
func newDB(setting config.Config) (IDB, error) {
	val := setting.Value("proto")
	proto := val.String()
	resolver, ok := dbResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("db: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(setting)
}
