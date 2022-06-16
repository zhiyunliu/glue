package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/zhiyunliu/glue/log"
)

var (
	_ Config = (*config)(nil)
)

func buildKeyNotFoundError(key string) error {
	return fmt.Errorf("key=%s not found", key)
}

// Observer is config observer.
type Observer func(string, Value)

// Config is a config interface.
type Config interface {
	Load() error
	Source(sources ...Source) error
	Scan(v interface{}) error
	Value(key string) Value
	Watch(key string, o Observer) error
	Close() error
	Get(key string) Config
	Root() Config
}

type config struct {
	opts      options
	reader    Reader
	cached    sync.Map
	observers sync.Map
	watchers  []Watcher
	//	log       log.Logger
}

// New new a config with options.
func New(opts ...Option) Config {
	o := options{
		logger:   log.DefaultLogger,
		decoder:  defaultDecoder,
		resolver: defaultResolver,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &config{
		opts:   o,
		reader: newReader(o),
	}
}

func (c *config) Get(key string) Config {
	return &wrap{
		rootConfig: c,
		curkey:     key,
	}
}

func (c *config) Root() Config {
	return c
}

func (c *config) watch(w Watcher) {
	for {
		kvs, err := w.Next()
		if errors.Is(err, context.Canceled) {
			c.opts.logger.Errorf("watcher's ctx cancel : %v", err)
			return
		}
		if errors.Is(err, ErrorUnchanged) {
			time.Sleep(time.Second)
			continue
		}
		if err != nil {
			time.Sleep(time.Second)
			c.opts.logger.Errorf("failed to watch next config: %v", err)
			continue
		}
		if err := c.reader.Merge(kvs...); err != nil {
			c.opts.logger.Errorf("failed to merge next config: %v", err)
			continue
		}
		if err := c.reader.Resolve(); err != nil {
			c.opts.logger.Errorf("failed to resolve next config: %v", err)
			continue
		}
		c.cached.Range(func(key, value interface{}) bool {
			k := key.(string)
			v := value.(Value)
			if n, ok := c.reader.Value(k); ok &&
				reflect.TypeOf(n.Load()) == reflect.TypeOf(v.Load()) &&
				!reflect.DeepEqual(n.Load(), v.Load()) {
				v.Store(n.Load())

				if o, ok := c.observers.Load(k); ok {
					o.(Observer)(k, v)
				}
			}
			return true
		})
	}
}

func (c *config) Source(sources ...Source) error {
	c.opts.sources = append(c.opts.sources, sources...)
	return c.loadSource(sources...)
}

func (c *config) Load() error {
	return c.loadSource(c.opts.sources...)
}

func (c *config) loadSource(sources ...Source) error {
	for _, src := range sources {
		kvs, err := src.Load()
		if err != nil {
			return err
		}
		for _, v := range kvs {
			c.opts.logger.Infof("config loaded: %s format: %s", v.Key, v.Format)
		}
		if err = c.reader.Merge(kvs...); err != nil {
			c.opts.logger.Errorf("failed to merge config source: %v", err)
			return err
		}
		w, err := src.Watch()
		if err != nil {
			c.opts.logger.Errorf("failed to watch config source: %v", err)
			return err
		}
		c.watchers = append(c.watchers, w)
		go c.watch(w)
	}
	if err := c.reader.Resolve(); err != nil {
		c.opts.logger.Errorf("failed to resolve config source: %v", err)
		return err
	}
	return nil
}

func (c *config) Value(key string) Value {
	if v, ok := c.cached.Load(key); ok {
		return v.(Value)
	}
	if v, ok := c.reader.Value(key); ok {
		c.cached.Store(key, v)
		return v
	}
	return &emptyValue{err: buildKeyNotFoundError(key)}
}

func (c *config) Scan(v interface{}) error {
	data, err := c.reader.Source()
	if err != nil {
		return err
	}
	return unmarshalJSON(data, v)
}

func (c *config) Watch(key string, o Observer) error {
	if v := c.Value(key); v.Load() == nil {
		return buildKeyNotFoundError(key)
	}
	c.observers.Store(key, o)
	return nil
}

func (c *config) Close() error {
	for _, w := range c.watchers {
		if err := w.Stop(); err != nil {
			return err
		}
	}
	return nil
}
