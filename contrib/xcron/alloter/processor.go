package alloter

import (
	sctx "context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/zhiyunliu/glue/contrib/alloter"
	"github.com/zhiyunliu/glue/dlocker"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/standard"
	"github.com/zhiyunliu/glue/xcron"
	"github.com/zhiyunliu/golibs/xstack"
)

// processor cron管理程序，用于管理多个任务的执行，暂停，恢复，动态添加，移除
type processor struct {
	ctx          sctx.Context
	lock         sync.Mutex
	closeChan    chan struct{}
	index        int
	jobs         cmap.ConcurrentMap
	monopolyJobs cmap.ConcurrentMap
	reqs         cmap.ConcurrentMap
	interval     time.Duration
	slots        [60]cmap.ConcurrentMap //time slots
	engine       *alloter.Engine
	onceLock     sync.Once
}

// NewProcessor 创建processor
func newProcessor(ctx sctx.Context, engine *alloter.Engine) (p *processor, err error) {
	p = &processor{
		ctx:          ctx,
		index:        -1,
		interval:     time.Second,
		closeChan:    make(chan struct{}),
		jobs:         cmap.New(),
		monopolyJobs: cmap.New(),
		reqs:         cmap.New(),
		engine:       engine,
	}

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
	defer ticker.Stop()
	for {
		select {
		case <-s.closeChan:
			return nil
		case tickTime := <-ticker.C:
			now := time.Now()
			if tickTime.Unix() > now.Unix() {
				log.Warnf("ticker quick than now.tick:%v,now:%v", tickTime.Unix(), now.Unix())
				// // 计算实际的时间间隔
				// actualInterval := now.Truncate(time.Second).Sub(now)

				// // 手动矫正下一次触发时间
				// nextTick := now.Add(time.Second - actualInterval)

				// // 输出当前时间和下一次触发时间
				// fmt.Println("Current time:", now)
				// fmt.Println("Next tick:", nextTick)

				// // 等待下一次触发时间
				// time.Sleep(nextTick.Sub(now))
			}
			s.index = (s.index + 1) % len(s.slots)
			go s.execute(s.index)
		}
	}
}

// Add 添加任务
func (s *processor) Add(jobs ...*xcron.Job) (err error) {
	for _, t := range jobs {
		if t.Disable {
			s.Remove(t.GetKey())
			continue
		}
		if err = t.Init(); err != nil {
			return
		}

		if err := s.checkMonopoly(t); err != nil {
			return err
		}
		req, err := newRequest(t)
		if err != nil {
			return fmt.Errorf("构建cron失败:cron=%s,service=%s,error:%v", t.Cron, t.Service, err)
		}
		req.ctx = s.ctx
		if err := s.reset(req); err != nil {
			return err
		}
	}
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

func (s *processor) checkMonopoly(j *xcron.Job) (err error) {
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
			locker: sdlocker.GetDLocker().Build(fmt.Sprintf("glue:cron:locker:%s", j.GetKey()), dlocker.WithData(fmt.Sprintf("c:%s,s:%s", j.Cron, j.Service))),
			expire: j.CalcExpireTime(),
		}
	})
	return nil
}

func (s *processor) reset(req *Request) (err error) {
	if req.job.Disable {
		return
	}
	req.reset()
	s.resetMonopolyJob(req.job)
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

func (s *processor) resetMonopolyJob(job *xcron.Job) {
	//根据执行后，重置下一次的独占时间
	if !job.IsMonopoly() {
		return
	}
	val, ok := s.monopolyJobs.Get(job.GetKey())
	if !ok {
		return
	}
	mjob := val.(*monopolyJob)
	mjob.expire = job.CalcExpireTime()

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

	return
}

func (s *processor) handle(req *Request) {
	defer func() {
		if obj := recover(); obj != nil {
			log.Panicf("cron.handle.Cron:%s,service:%s, error:%+v. stack:%s", req.job.Cron, req.job.Service, obj, xstack.GetStack(1))
		}
	}()

	rangeSecs := time.Since(req.job.CalcNextTime).Seconds()
	//时间差距超过1分钟
	if math.Abs(rangeSecs) >= 60 {
		log.Warnf("cron.handle.Cron.1:%s,service:%s,over 60s.calc:%d,now:%d", req.job.Cron, req.job.Service, req.job.CalcNextTime.Unix(), time.Now().Unix())
		s.reset(req)
		return
	}

	hasMonopoly, err := req.Monopoly(s.monopolyJobs)
	if err != nil {
		log.Warnf("cron.handle.Cron.2:%s,service:%s, error:%+v. stack:%s", req.job.Cron, req.job.Service, err, xstack.GetStack(1))
		s.reset(req)
		return
	}
	if hasMonopoly {
		log.Warnf("cron.handle.Cron.3:%s,service:%s,meta:%+v=>monopoly.key=%s", req.job.Cron, req.job.Service, req.job.Meta, req.job.GetKey())
		s.reset(req)
		return
	}

	req.ctx = sctx.Background()
	resp := newResponse()
	err = s.engine.HandleRequest(req, resp)
	if err != nil {
		panic(err)
	}
	resp.Flush()
	s.reset(req)
}

func (s *processor) execute(idx int) {
	current := s.slots[idx]
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
	job    *xcron.Job
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
