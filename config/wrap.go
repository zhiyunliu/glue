package config

import "fmt"

type wrap struct {
	curkey     string
	rootConfig Config
}

func (l *wrap) Load() error {
	return nil
}
func (l *wrap) Scan(v interface{}) error {
	return l.rootConfig.Value(l.curkey).Scan(v)
}
func (l *wrap) Value(key string) Value {
	return l.rootConfig.Value(fmt.Sprintf("%s.%s", l.curkey, key))
}
func (l *wrap) Watch(key string, o Observer) error {
	return nil
}
func (l *wrap) Close() error {
	return nil
}
func (l *wrap) Get(key string) Config {
	return &wrap{
		rootConfig: l.rootConfig,
		curkey:     fmt.Sprintf("%s.%s", l.curkey, key),
	}
}
