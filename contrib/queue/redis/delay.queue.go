package redis

import (
	"context"

	"github.com/zhiyunliu/glue/contrib/queue/redisdelay"
	"github.com/zhiyunliu/glue/queue"
)

func (p *Producer) appendDelay(ctx context.Context, orgQueue string, msg queue.Message, delaySeconds int64) (err error) {

	tmpProcessor, ok := p.delayQueueMap.Load(orgQueue)
	if !ok {
		actual, loaded := p.delayQueueMap.LoadOrStore(orgQueue, redisdelay.NewProcessor(p.client, orgQueue, p.opts.DelayInterval, p.BatchPush))
		if !loaded {
			if processor, ok := actual.(queue.DelayProcessor); ok {
				processor.Start(p.closeChan)
			}
		}
		tmpProcessor = actual
	}
	return tmpProcessor.(queue.DelayProcessor).AppendMessage(ctx, msg, delaySeconds)
}

func (p *Producer) BatchPush(ctx context.Context, key string, msgList ...queue.Message) error {
	if len(msgList) == 0 {
		return nil
	}
	for i := range msgList {
		if err := p.Push(ctx, key, msgList[i]); err != nil {
			return err
		}
	}
	return nil
}
