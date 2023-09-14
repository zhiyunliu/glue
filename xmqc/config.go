package xmqc

import (
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/metadata"
)

type Config struct {
	Addr   string        `json:"addr"`
	Status engine.Status `json:"status"`
	Proto  string        `json:"proto"`
}

type Task struct {
	Queue       string            `json:"queue"`
	Service     string            `json:"service,omitempty"`
	Disable     bool              `json:"disable"`
	Concurrency int               `json:"concurrency,omitempty"`
	Meta        metadata.Metadata `json:"meta,omitempty"`
}

type TaskList []*Task

func (t *Task) GetQueue() string {
	return t.Queue
}

func (t *Task) GetConcurrency() int {
	return t.Concurrency
}

func (t *Task) GetService() string {
	if t.Service != "" {
		return t.Service
	}
	t.Service = GetService(t.Queue)
	return t.Service
}
