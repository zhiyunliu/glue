package redisdelay

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	rds "github.com/redis/go-redis/v9"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
	"golang.org/x/sync/errgroup"
)

var _noneDelayData = errors.New("none delay data")

//go:embed delay.lua
var DelayProcessScript string

type delayProcess struct {
	client        *redis.Client
	callback      queue.DelayCallback
	orgQueue      string
	delayQueue    string
	scriptHash    string
	delayInterval int
	groups        *errgroup.Group
}

func NewProcessor(client *redis.Client, orgQueue string, delayInterval int, callback queue.DelayCallback) queue.DelayProcessor {

	return &delayProcess{
		client:        client,
		callback:      callback,
		orgQueue:      orgQueue,
		delayQueue:    fmt.Sprintf("%s:delay", orgQueue),
		delayInterval: delayInterval,
		groups:        &errgroup.Group{},
	}

}

func (p delayProcess) Start(done chan struct{}) error {
	p.groups.Go(func() error {
		delayInterval := time.Second * time.Duration(p.delayInterval)

		for {
			select {
			case <-done:
				return nil
			default:
				msgList, err := p.procDelayQueue()
				if err == _noneDelayData {
					time.Sleep(delayInterval)
					continue
				}
				if err != nil {
					log.Error("redisdelay.procDelayQueue", p.orgQueue, err)
					time.Sleep(delayInterval)
					continue
				}
				err = p.callback(context.Background(), p.orgQueue, msgList...)
				if err != nil {
					log.Error("redisdelay.delayProcess", p.orgQueue, err)
				}
			}
		}
	})
	return nil
}

func (p delayProcess) AppendMessage(ctx context.Context, msg queue.Message, delaySeconds int64) (err error) {
	newScore := time.Now().Unix() + delaySeconds
	err = p.client.ZAdd(ctx, p.delayQueue, rds.Z{Score: float64(newScore), Member: msg}).Err()
	return
}

func (p *delayProcess) procDelayQueue() (msgList []queue.Message, err error) {
	if p.scriptHash == "" {
		p.scriptHash, err = p.client.ScriptLoad(context.Background(), DelayProcessScript).Result()
		if err != nil {
			err = fmt.Errorf("ScriptLoad:%s,err:%+v", DelayProcessScript, err)
			return
		}
	}
	tmpValList, err := p.client.EvalSha(context.Background(), p.scriptHash, []string{p.delayQueue}, time.Now().Unix()).Result()
	if err != nil {
		err = fmt.Errorf("EvalSha:%s,err:%+v", DelayProcessScript, err)
		return
	}
	valList := tmpValList.([]any)
	if len(valList) == 0 {
		return nil, _noneDelayData
	}

	msgList = make([]queue.Message, 0, len(valList))

	for i := range valList {
		val := valList[i].(string)

		msgItem := &queue.MsgItem{
			HeaderMap: make(xtypes.SMap),
		}
		//可能会丢消息
		err = json.Unmarshal(bytesconv.StringToBytes(val), &msgItem)
		if err != nil {
			log.Errorf("queue:%s,Unmarshal:%s,err:%+v", p.orgQueue, val, err)
			continue
		}
		msgItem.HeaderMap[constants.HeaderSourceIp] = global.LocalIp
		msgItem.HeaderMap[constants.HeaderSourceName] = global.AppName
		msgList = append(msgList, msgItem)
	}

	return msgList, nil
}
