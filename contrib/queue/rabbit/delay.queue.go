package rabbit

import (
	"fmt"

	"github.com/zhiyunliu/glue/queue"
	"golang.org/x/sync/errgroup"
)

func (p *Producer) appendDelay(orgQueue string, msg queue.Message, delaySeconds int64) (err error) {

	tmpProcessor, ok := p.delayQueueMap.Load(orgQueue)
	if !ok {
		actual, loaded := p.delayQueueMap.LoadOrStore(orgQueue, p.newProcessor(orgQueue, p.BatchPush))
		if !loaded {
			if processor, ok := actual.(queue.DelayProcessor); ok {
				processor.Start(p.closeChan)
			}
		}
		tmpProcessor = actual
	}
	return tmpProcessor.(queue.DelayProcessor).AppendMessage(msg, delaySeconds)
}

func (p *Producer) BatchPush(key string, msgList ...queue.Message) error {
	if len(msgList) == 0 {
		return nil
	}
	for i := range msgList {
		if err := p.Push(key, msgList[i]); err != nil {
			return err
		}
	}
	return nil
}

func (p *Producer) newProcessor(orgQueue string, callback queue.DelayCallback) queue.DelayProcessor {

	return &delayProcess{
		callback:   callback,
		orgQueue:   orgQueue,
		delayQueue: fmt.Sprintf("%s:delay", orgQueue),
		groups:     &errgroup.Group{},
	}
}

type delayProcess struct {
	callback   queue.DelayCallback
	orgQueue   string
	delayQueue string
	groups     *errgroup.Group
}

func (p delayProcess) Start(done chan struct{}) {

}

func (p delayProcess) AppendMessage(msg queue.Message, delaySeconds int64) (err error) {

	return
}
