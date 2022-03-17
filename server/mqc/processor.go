package mqc

import (
	"fmt"
	"sync"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/contrib/alloter"
	"github.com/zhiyunliu/velocity/queue"
	"github.com/zhiyunliu/velocity/server"
)

//processor cron管理程序，用于管理多个任务的执行，暂停，恢复，动态添加，移除
type processor struct {
	lock      sync.Mutex
	closeChan chan struct{}
	queues    cmap.ConcurrentMap
	consumer  queue.IMQC
	status    server.RunStatus
	engine    *alloter.Engine
	onceLock  sync.Once
}

//NewProcessor 创建processor
func newProcessor(setting config.Config) (p *processor, err error) {
	p = &processor{
		status:    server.Unstarted,
		closeChan: make(chan struct{}),
		queues:    cmap.New(),
	}

	p.consumer, err = queue.NewMQC(setting)
	if err != nil {
		return nil, fmt.Errorf("构建mqc服务失败(raw:%s) %v", setting.String(), err)
	}
	p.engine = alloter.New()

	return p, nil
}

//QueueItems QueueItems
func (s *processor) QueueItems() map[string]interface{} {
	return s.queues.Items()
}

//Start 所有任务
func (s *processor) Start() error {
	if err := s.consumer.Connect(); err != nil {
		return err
	}
	_, err := s.Resume()
	return err
}

//Add 添加队列信息
func (s *processor) Add(tasks ...*Task) error {
	for _, task := range tasks {
		if ok := s.queues.SetIfAbsent(task.Queue, task); ok && s.status == server.Running {
			if err := s.consume(task); err != nil {
				return err
			}
		}
	}
	return nil
}

//Remove 除移队列信息
func (s *processor) Remove(tasks ...*Task) error {
	for _, t := range tasks {
		s.consumer.Unconsume(t.Queue)
		s.queues.Remove(t.Queue)
	}
	return nil
}

//Pause 暂停所有任务
func (s *processor) Pause() (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.status != server.Pause {
		s.status = server.Pause
		items := s.queues.Items()
		for _, v := range items {
			queue := v.(*Task)
			s.consumer.Unconsume(queue.Queue) //取消服务订阅
		}
		return true, nil
	}
	return false, nil
}

//Resume 恢复所有任务
func (s *processor) Resume() (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.status != server.Running {
		s.status = server.Running
		items := s.queues.Items()
		for _, v := range items {
			queue := v.(*Task)
			if err := s.consume(queue); err != nil {
				return true, err
			}
		}
		return true, nil
	}
	return false, nil
}
func (s *processor) consume(task *Task) error {
	return s.consumer.Consume(task.Queue, s.handleCallback(task))
}

//Close 退出
func (s *processor) Close() error {
	s.onceLock.Do(func() {
		close(s.closeChan)
		s.Pause()
	})
	return nil
}

func (s *processor) handleCallback(task *Task) func(queue.IMQCMessage) {
	return func(m queue.IMQCMessage) {
		req, err := NewRequest(task, m)
		if err != nil {
			panic(err)
		}
		writer, err := s.engine.HandleRequest(req)
		if err != nil {
			panic(err)
		}
		writer.Flush()
	}
}
