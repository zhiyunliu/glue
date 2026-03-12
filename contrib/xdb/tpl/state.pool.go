package tpl

import (
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

var _ xdb.SqlStatePool = &StatePool{}

type StatePool struct {
	pool *sync.Pool
}

func NewStatePool(newFunc func() any) xdb.SqlStatePool {
	return &StatePool{
		pool: &sync.Pool{
			New: newFunc,
		},
	}
}
func (sp *StatePool) Get() xdb.SqlState {
	return sp.pool.Get().(xdb.SqlState)
}

func (sp *StatePool) Put(state xdb.SqlState) {
	sp.pool.Put(state)
}
