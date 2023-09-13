package alloter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/metadata"
	"github.com/zhiyunliu/glue/server"
)

type clientConfig struct {
	Name         string          `json:"-"`
	ConnTimeout  int             `json:"conn_timeout"`
	Balancer     string          `json:"balancer"`      //负载类型 round_robin:论寻负载
	ServerConfig json.RawMessage `json:"server_config"` //
	Trace        bool            `json:"trace"`
	Config       config.Config   `json:"-"`
}

type serverConfig struct {
	Config Config   `json:"config" yaml:"config"`
	Tasks  TaskList `json:"tasks"  yaml:"tasks"`
}

type Config struct {
	Addr   string        `json:"addr"`
	Status server.Status `json:"status"`
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
	t.Service = getService(t.Queue)
	return t.Service
}

func getService(queue string) string {
	if strings.HasPrefix(queue, "/") {
		return queue
	}
	tmp := queue
	tmp = strings.ReplaceAll(tmp, ":", "_")
	return fmt.Sprintf("/mqc_%s", tmp)
}
