package robfigcron

import (
	sctx "context"
	"fmt"
	"sync"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/robfig/cron/v3"
	"github.com/zhiyunliu/glue/contrib/alloter"
	"github.com/zhiyunliu/glue/dlocker"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/standard"
	"github.com/zhiyunliu/glue/xcron"
	"github.com/zhiyunliu/golibs/xstack"
)

// processor cron管理程序，用于管理多个任务的执行，暂停，恢复，动态添加，移除
type processor struct {
	ctx           sctx.Context
	closeChan     chan struct{}
	jobs          cmap.ConcurrentMap
	monopolyJobs  cmap.ConcurrentMap
	engine        *alloter.Engine
	onceLock      sync.Once
	cronStdEngine *cron.Cron
	cronSecEngine *cron.Cron
}

type procJob struct {
	job     *xcron.Job
	entryid cron.EntryID
	engine  *cron.Cron
}

// NewProcessor 创建processor
func newProcessor(ctx sctx.Context, engine *alloter.Engine) (p *processor, err error) {
	p = &processor{
		ctx:           ctx,
		closeChan:     make(chan struct{}),
		jobs:          cmap.New(),
		monopolyJobs:  cmap.New(),
		engine:        engine,
		cronStdEngine: cron.New(),
		cronSecEngine: cron.New(cron.WithSeconds()),
	}
	return p, nil
}

// Items Items
func (s *processor) Items() map[string]interface{} {
	return s.jobs.Items()
}

// Start 所有任务
func (s *processor) Start() error {
	go s.cronStdEngine.Run()
	go s.cronSecEngine.Run()
	return nil
}

// Add 添加任务
func (s *processor) Add(jobs ...*xcron.Job) (err error) {
	var curEngine *cron.Cron
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
		var jobId cron.EntryID
		curEngine = s.cronStdEngine
		if t.WithSeconds {
			curEngine = s.cronSecEngine
		}
		jobId, err = curEngine.AddJob(t.Cron, s.buildFuncJob(t))
		if err != nil {
			return err
		}
		s.jobs.Set(t.GetKey(), &procJob{
			job:     t,
			entryid: jobId,
			engine:  curEngine,
		})
	}
	return
}

func (s *processor) buildFuncJob(job *xcron.Job) cron.FuncJob {
	return func() {
		req := newRequest(job)
		req.ctx = s.ctx
		s.handle(req)
	}
}

// Remove 移除服务
func (s *processor) Remove(key string) {
	if req, ok := s.jobs.Get(key); ok {
		procJob := req.(*procJob)
		procJob.job.Disable = true
		procJob.engine.Remove(procJob.entryid)
	}
	s.jobs.Remove(key)
}

// Close 退出
func (s *processor) Close() error {
	s.onceLock.Do(func() {
		close(s.closeChan)
		s.cronStdEngine.Stop()
		s.cronSecEngine.Stop()
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
			locker: sdlocker.GetDLocker().Build(fmt.Sprintf("glue:cron:locker:%s", j.GetKey()), dlocker.WithData(j.GetLockData())),
			expire: j.CalcExpireSeconds(),
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
	mjob.expire = job.CalcExpireSeconds()

}

func (s *processor) closeMonopolyJobs() {
	for item := range s.monopolyJobs.IterBuffered() {
		item.Val.(*monopolyJob).Close()
	}
	s.monopolyJobs.Clear()
}

func (s *processor) handle(req *Request) {
	defer func() {
		if obj := recover(); obj != nil {
			log.Panicf("cron.handle.Cron:%s,service:%s, error:%+v. stack:%s", req.job.Cron, req.job.Service, obj, xstack.GetStack(1))
		}
	}()

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
	req.header["x-cron-job-key"] = req.job.GetKey()
	req.ctx = sctx.Background()
	resp := newResponse()
	err = s.engine.HandleRequest(req, resp)
	if err != nil {
		panic(err)
	}
	resp.Flush()
	s.reset(req)
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
