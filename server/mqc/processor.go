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

//Processor cron管理程序，用于管理多个任务的执行，暂停，恢复，动态添加，移除
type Processor struct {
	lock      sync.Mutex
	closeChan chan struct{}
	queues    cmap.ConcurrentMap
	consumer  queue.IMQC
	status    server.RunStatus
	engine    *alloter.Engine
}

//NewProcessor 创建processor
func NewProcessor(setting config.Config) (p *Processor, err error) {
	p = &Processor{
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
func (s *Processor) QueueItems() map[string]interface{} {
	return s.queues.Items()
}

//Start 所有任务
func (s *Processor) Start(wait ...bool) error {
	if err := s.consumer.Connect(); err != nil {
		return err
	}
	if len(wait) > 0 && !wait[0] {
		_, err := s.Resume()
		return err
	}
	return nil
}

//Add 添加队列信息
func (s *Processor) Add(tasks ...*Task) error {
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
func (s *Processor) Remove(tasks ...*Task) error {
	for _, t := range tasks {
		s.consumer.Unconsume(t.Queue)
		s.queues.Remove(t.Queue)
	}
	return nil
}

//Pause 暂停所有任务
func (s *Processor) Pause() (bool, error) {
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
func (s *Processor) Resume() (bool, error) {
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
func (s *Processor) consume(task *Task) error {
	if err := s.consumer.Consume(task.Queue, s.handleCallback(task)); err != nil {
		return err
	}
	return nil
}

//Close 退出
func (s *Processor) Close() {

}

func (s *Processor) handleCallback(task *Task) func(queue.IMQCMessage) {
	return func(m queue.IMQCMessage) {
		req, err := NewRequest(task, m)
		if err != nil {
			panic(err)
		}
		s.engine.HandleRequest(req)
	}
}
