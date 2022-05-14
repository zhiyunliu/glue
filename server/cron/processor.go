package cron

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/zhiyunliu/gel/contrib/alloter"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/gel/server"
	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xstack"
)

//processor cron管理程序，用于管理多个任务的执行，暂停，恢复，动态添加，移除
type processor struct {
	lock      sync.Mutex
	closeChan chan struct{}
	index     int
	jobs      cmap.ConcurrentMap
	interval  time.Duration
	slots     [60]cmap.ConcurrentMap //time slots
	status    server.RunStatus
	engine    *alloter.Engine
	onceLock  sync.Once
}

//NewProcessor 创建processor
func newProcessor() (p *processor, err error) {
	p = &processor{
		interval:  time.Second,
		status:    server.Unstarted,
		closeChan: make(chan struct{}),
		jobs:      cmap.New(),
	}

	p.engine = alloter.New()

	for i := range p.slots {
		p.slots[i] = cmap.New()
	}
	return p, nil
}

//Items Items
func (s *processor) Items() map[string]interface{} {
	return s.jobs.Items()
}

//Start 所有任务
func (s *processor) Start() error {
	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-s.closeChan:
			return nil
		case <-ticker.C:
			s.execute()
		}
	}
}

//Add 添加任务
func (s *processor) Add(jobs ...*Job) (err error) {
	for _, t := range jobs {
		if t.Disable {
			s.Remove(t.GetKey())
			continue
		}
		req, err := NewRequest(t)
		if err != nil {
			return fmt.Errorf("构建cron失败:cron=%s,service=%s,error:%v", t.Cron, t.Service, err)
		}

		if err := s.reset(req); err != nil {
			return err
		}
	}
	return

}

func (s *processor) reset(req *Request) (err error) {
	req.reset()
	now := time.Now()
	nextTime := req.NextTime(now)
	if nextTime.Sub(now) < 0 {
		return errors.New("next time less than now.1")
	}
	offset, round := s.getOffset(now, nextTime)
	req.round.Update(round)
	s.slots[offset].Set(session.Create(), req)
	return
}

//Remove 移除服务
func (s *processor) Remove(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, slot := range s.slots {
		slot.Remove(key)
	}
}

//Close 退出
func (s *processor) Close() error {
	s.onceLock.Do(func() {
		close(s.closeChan)
	})
	return nil
}

func (s *processor) getOffset(now time.Time, next time.Time) (pos int, circle int) {
	secs := next.Sub(now).Seconds() //剩余时间
	delaySeconds := int(math.Ceil(secs))
	circle = int(delaySeconds) / len(s.slots)
	pos = int(s.index+delaySeconds) % len(s.slots)
	if pos == s.index { //offset与当前index相同时，应减少一环
		circle--
	}
	return
}

func (s *processor) handle(req *Request) {
	defer func() {
		if obj := recover(); obj != nil {
			log.Panicf("cron.handle.Cron:%s,service:%s, error:%+v. stack:%s", req.job.Cron, req.job.Service, obj, xstack.GetStack(1))
		}
	}()

	resp, err := NewResponse(req.job)
	if err != nil {
		panic(err)
	}

	err = s.engine.HandleRequest(req, resp)
	if err != nil {
		panic(err)
	}
	resp.Flush()
}

func (s *processor) execute() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.index = (s.index + 1) % len(s.slots)
	current := s.slots[s.index]

	removeKeyList := []string{}
	resetJobList := []*Request{}

	current.IterCb(func(key string, value interface{}) {
		jobReq := value.(*Request)
		curidx := jobReq.round.Current()
		if curidx > 0 {
			jobReq.round.Reduce()
			return
		}
		go s.handle(jobReq)
		removeKeyList = append(removeKeyList, key)
		resetJobList = append(resetJobList, jobReq)
	})

	for _, key := range removeKeyList {
		current.Remove(key)
	}

	for _, jobReq := range resetJobList {
		s.reset(jobReq)
	}
}
