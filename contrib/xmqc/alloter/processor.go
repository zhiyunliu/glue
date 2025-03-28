package alloter

import (
	"context"
	"fmt"
	"sync"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/zhiyunliu/alloter"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/glue/xmqc"
	"github.com/zhiyunliu/golibs/xstack"
)

// processor cron管理程序，用于管理多个任务的执行，暂停，恢复，动态添加，移除
type processor struct {
	ctx        context.Context
	lock       sync.Mutex
	closeChan  chan struct{}
	queues     cmap.ConcurrentMap[string, *xmqc.Task]
	consumer   queue.IMQC
	status     engine.RunStatus
	engine     *alloter.Engine
	onceLock   sync.Once
	configName string
}

// NewProcessor 创建processor
func newProcessor(ctx context.Context, alloterEngine *alloter.Engine, proto, configName string, setting config.Config) (p *processor, err error) {
	p = &processor{
		ctx:        ctx,
		status:     engine.Unstarted,
		closeChan:  make(chan struct{}),
		queues:     cmap.New[*xmqc.Task](),
		engine:     alloterEngine,
		configName: configName,
	}

	p.consumer, err = queue.NewMQC(proto, configName, setting)
	if err != nil {
		return nil, fmt.Errorf("构建mqc服务失败:%v", err)
	}
	return p, nil
}

// QueueItems QueueItems
func (s *processor) QueueItems() map[string]*xmqc.Task {
	return s.queues.Items()
}

// Start 所有任务
func (s *processor) Start() error {
	if err := s.consumer.Connect(); err != nil {
		return err
	}
	_, err := s.Resume()
	if err != nil {
		return err
	}
	return s.consumer.Start()
}

// Add 添加队列信息
func (s *processor) Add(tasks ...*xmqc.Task) error {
	for _, task := range tasks {
		if task.Disable {
			continue
		}
		if ok := s.queues.SetIfAbsent(task.Queue, task); ok && s.status == engine.Running {
			if err := s.consume(task); err != nil {
				return err
			}
		}
	}
	return nil
}

// Remove 除移队列信息
func (s *processor) Remove(tasks ...*xmqc.Task) error {
	for _, t := range tasks {
		s.consumer.Unconsume(t.Queue)
		s.queues.Remove(t.Queue)
	}
	return nil
}

// Pause 暂停所有任务
func (s *processor) Pause() (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.status != engine.Pause {
		s.status = engine.Pause
		items := s.queues.Items()
		for _, v := range items {
			s.consumer.Unconsume(v.Queue) //取消服务订阅
		}
		return true, nil
	}
	return false, nil
}

// Resume 恢复所有任务
func (s *processor) Resume() (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.status != engine.Running {
		s.status = engine.Running
		items := s.queues.Items()
		for _, v := range items {
			err := func(tsk *xmqc.Task) (ierr error) {
				return s.consume(tsk)
			}(v)
			if err != nil {
				return true, err
			}
		}
		return true, nil
	}
	return false, nil
}
func (s *processor) consume(task *xmqc.Task) error {
	task.FullPath = fmt.Sprint(s.consumer.ServerURL(), task.GetService())
	return s.consumer.Consume(task, s.handleCallback(task))
}

// Close 退出
func (s *processor) Close() error {
	s.onceLock.Do(func() {
		close(s.closeChan)
		s.Pause()
	})
	return nil
}

func (s *processor) handleCallback(task *xmqc.Task) func(queue.IMQCMessage) {
	return func(m queue.IMQCMessage) {
		defer func() {
			if obj := recover(); obj != nil {
				log.Panicf("mqc.handleCallback.Queue:%s,data:%s, error:%+v. stack:%s", task.Queue, m.Original(), obj, xstack.GetStack(1))
			}
		}()

		req := newRequest(task, m)
		req.ctx = context.Background()
		resp := newResponse(task, m)

		err := s.engine.HandleRequest(req, resp)
		if err != nil {
			m.Nack(err)
			panic(err)
		}
	}
}
