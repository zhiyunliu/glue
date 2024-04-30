package xdb

type DbParam interface {
	ToDbParam() map[string]any
}
