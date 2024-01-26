package redisdelay

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	rds "github.com/go-redis/redis/v7"

	"github.com/zhiyunliu/glue/contrib/redis"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/golibs/bytesconv"
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

func (p delayProcess) Start(done chan struct{}) {
	p.groups.Go(func() error {
		delayInterval := time.Second * time.Duration(p.delayInterval)

		for {
			select {
			case <-done:
				return nil
			default:
				msg, err := p.procDelayQueue()
				if err == _noneDelayData {
					time.Sleep(delayInterval)
					continue
				}
				if err != nil {
					log.Error("redisdelay.procDelayQueue", p.orgQueue, err)
					time.Sleep(delayInterval)
					continue
				}
				err = p.callback(p.orgQueue, msg)
				if err != nil {
					log.Error("redisdelay.delayProcess", p.orgQueue, err)
				}
			}
		}
	})
}

func (p delayProcess) AppendMessage(msg queue.Message, delaySeconds int64) (err error) {

	bytes, _ := json.Marshal(map[string]interface{}{
		"header": msg.Header(),
		"body":   msg.Body(),
	})

	newScore := time.Now().Unix() + delaySeconds
	err = p.client.ZAdd(p.delayQueue, &rds.Z{Score: float64(newScore), Member: string(bytes)}).Err()
	return
}

func (p *delayProcess) procDelayQueue() (msg queue.Message, err error) {
	if p.scriptHash == "" {
		p.scriptHash, err = p.client.ScriptLoad(DelayProcessScript).Result()
		if err != nil {
			err = fmt.Errorf("ScriptLoad:%s,err:%+v", DelayProcessScript, err)
			return
		}
	}
	tmpValList, err := p.client.EvalSha(p.scriptHash, []string{p.delayQueue}, time.Now().Unix()).Result()
	if err != nil {
		err = fmt.Errorf("EvalSha:%s,err:%+v", DelayProcessScript, err)
		return
	}
	valList := tmpValList.([]any)
	if len(valList) == 0 {
		return nil, _noneDelayData
	}
	val := valList[0].(string)

	msgItem := &queue.MsgItem{}
	err = json.Unmarshal(bytesconv.StringToBytes(val), &msgItem)
	if err != nil {
		err = fmt.Errorf("Unmarshal:%s,err:%+v", val, err)
	}
	return msgItem, err
}
