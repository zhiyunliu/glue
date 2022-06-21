package demos

import (
	"time"

	gel "github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/context"
)

type Dlockdemo struct{}

func NewDLock() *Dlockdemo {
	return &Dlockdemo{}
}

func (d *Dlockdemo) CreateHandle(ctx context.Context) interface{} {
	lockKey := ctx.Request().Query().Get("key")
	if lockKey == "" {
		lockKey = "lockkey"
	}
	locker := gel.DLocker(lockKey)
	isok, err := locker.Acquire(10)
	if err != nil {
		ctx.Log().Errorf("Acquire err:%+v", err)
		return err
	}
	if isok {
		sleeptime := 8
		ctx.Log().Debugf("sleep %d seconds", sleeptime)
		time.Sleep(time.Duration(sleeptime) * time.Second)
		err = locker.Renewal(10)
		ctx.Log().Debugf("Renewal 10 seconds %+v", err)
		return "success"
	}
	return "lock failure"
}
