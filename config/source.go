package config

// KeyValue is config key value.
type KeyValue struct {
	Key    string
	Value  []byte
	Format string
}

// Source is config source.
type Source interface {
	Name() string
	Load() ([]*KeyValue, error)
	Watch() (Watcher, error)
	Path() string
}

// Watcher watches a source for changes.
type Watcher interface {
	Next() ([]*KeyValue, error)
	Stop() error
}
