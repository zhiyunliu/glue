package cron

import (
	"context"
	sctx "context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	cron "github.com/robfig/cron/v3"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/contrib/alloter"
	"github.com/zhiyunliu/glue/dlocker"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/server"
	"github.com/zhiyunliu/glue/standard"
	"github.com/zhiyunliu/golibs/xstack"
)

// processor cron管理程序，用于管理多个任务的执行，暂停，恢复，动态添加，移除
type processor struct {
	ctx          context.Context
	lock         sync.Mutex
	closeChan    chan struct{}
	index        int
	jobs         cmap.ConcurrentMap
	monopolyJobs cmap.ConcurrentMap
	reqs         cmap.ConcurrentMap
	interval     time.Duration
	slots        [60]cmap.ConcurrentMap //time slots
	status       server.RunStatus
	engine       *alloter.Engine
	onceLock     sync.Once
	cfg          config.Config
}

// NewProcessor 创建processor
func newProcessor(cfg config.Config) (p *processor, err error) {
	p = &processor{
		index:        -1,
		interval:     time.Second,
		status:       server.Unstarted,
		closeChan:    make(chan struct{}),
		jobs:         cmap.New(),
		monopolyJobs: cmap.New(),
		reqs:         cmap.New(),
	}

	p.engine = alloter.New()

	for i := range p.slots {
		p.slots[i] = cmap.New()
	}
	return p, nil
}

// Items Items
func (s *processor) Items() map[string]interface{} {
	return s.jobs.Items()
}

// Start 所有任务
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

// Add 添加任务
func (s *processor) Add(jobs ...*Job) (err error) {
	for _, t := range jobs {
		if t.Disable {
			s.Remove(t.GetKey())
			continue
		}
		parser := cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		)
		if t.WithSeconds {
			parser = cron.NewParser(
				cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
			)
		}
		t.schedule, err = parser.Parse(t.Cron)
		if err != nil {
			return
		}
		if err := s.checkMonopoly(t); err != nil {
			return err
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

func (s *processor) checkMonopoly(j *Job) (err error) {
	if !j.IsMonopoly() {
		return nil
	}
	defer func() {
		if obj := recover(); obj != nil {
			err = fmt.Errorf("cron任务包含monopoly时需要提供dlocker的配置:%v", obj)

		}
	}()
	ins := standard.GetInstance(dlocker.TypeNode)
	sdlocker := ins.(dlocker.StandardLocker)
	s.monopolyJobs.Upsert(j.GetKey(), j, func(exist bool, valueInMap, newValue interface{}) interface{} {
		if exist {
			return valueInMap
		}
		return &monopolyJob{
			job:    j,
			locker: sdlocker.GetDLocker().Build(fmt.Sprintf("glue:cron:locker:%s", j.GetKey())),
			expire: int(math.Ceil(j.schedule.Next(time.Now()).Sub(time.Now()).Seconds())),
		}
	})
	return nil
}

func (s *processor) reset(req *Request) (err error) {
	if req.job.Disable {
		return
	}
	req.reset()
	now := time.Now()
	nextTime := req.job.NextTime(now)
	if nextTime.Sub(now) < 0 {
		return errors.New("next time less than now.1")
	}
	offset, round := s.getOffset(now, nextTime)
	req.round.Update(round)
	s.slots[offset].Set(req.session, req)
	s.reqs.Set(req.job.GetKey(), req)
	return
}

// Remove 移除服务
func (s *processor) Remove(key string) {
	if req, ok := s.reqs.Get(key); ok {
		req.(*Request).job.Disable = true
	}
	s.reqs.Remove(key)
}

// Close 退出
func (s *processor) Close() error {
	s.onceLock.Do(func() {
		close(s.closeChan)
		s.closeMonopolyJobs()
	})
	return nil
}

func (s *processor) closeMonopolyJobs() {
	for item := range s.monopolyJobs.IterBuffered() {
		item.Val.(*monopolyJob).Close()
	}
	s.monopolyJobs.Clear()
}

func (s *processor) getOffset(now time.Time, next time.Time) (pos int, circle int) {
	// 立即执行的任务放在下一秒执行
	if now == next {
		return s.index + 1, 0
	}
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
	hasMonopoly, err := req.Monopoly(s.monopolyJobs)
	if err != nil {
		log.Panicf("cron.handle.Cron.2:%s,service:%s, error:%+v. stack:%s", req.job.Cron, req.job.Service, err, xstack.GetStack(1))
		s.reset(req)
		return
	}
	if hasMonopoly {
		s.reset(req)
		return
	}

	req.ctx = sctx.Background()
	resp := NewResponse()
	err = s.engine.HandleRequest(req, resp)
	if err != nil {
		panic(err)
	}
	resp.Flush()
	s.reset(req)
}

func (s *processor) execute() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.index = (s.index + 1) % len(s.slots)
	current := s.slots[s.index]

	resetJobList := []*Request{}

	current.IterCb(func(key string, value interface{}) {
		jobReq := value.(*Request)
		if !jobReq.round.CanProc() {
			return
		}
		if jobReq.CanProc() && !jobReq.job.Disable {
			go s.handle(jobReq)
		}

		resetJobList = append(resetJobList, jobReq)
	})

	for _, jobReq := range resetJobList {
		current.Remove(jobReq.session)
	}
}

type monopolyJob struct {
	job    *Job
	locker dlocker.DLocker
	expire int
}

func (j *monopolyJob) Acquire() (bool, error) {
	return j.locker.Acquire(j.expire)
}

func (j *monopolyJob) Renewal() {
	j.locker.Renewal(j.expire)
}

func (j *monopolyJob) Close() {
	j.locker.Release()
}
