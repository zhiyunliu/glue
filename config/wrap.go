package config

import "fmt"

type wrap struct {
	curkey     string
	rootConfig Config
}

func (l wrap) Load() error {
	return nil
}

// Deprecated: As of Go v0.5.3, this function simply calls [ScanTo].
func (l wrap) Scan(v interface{}) error {
	return l.ScanTo(v)
}

func (l wrap) ScanTo(v interface{}) error {
	return l.rootConfig.Value(l.curkey).Scan(v)
}

func (l wrap) Value(key string) Value {
	return l.rootConfig.Value(fmt.Sprintf("%s.%s", l.curkey, key))
}

func (l *wrap) Watch(key string, o Observer) error {
	return l.rootConfig.Watch(fmt.Sprintf("%s.%s", l.curkey, key), o)
}

func (l wrap) Close() error {
	return nil
}

func (l wrap) Get(key string) Config {
	return &wrap{
		rootConfig: l.rootConfig,
		curkey:     fmt.Sprintf("%s.%s", l.curkey, key),
	}
}

func (l wrap) Path() string {
	return l.curkey
}

func (c wrap) Root() Config {
	return c.rootConfig
}

func (c *wrap) Source(sources ...Source) error {
	return c.rootConfig.Source(sources...)
}
