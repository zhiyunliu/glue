package redis

import (
	"github.com/zhiyunliu/glue/contrib/queue/redisdelay"
	"github.com/zhiyunliu/glue/queue"
)

func (p *Producer) appendDelay(orgQueue string, msg queue.Message, delaySeconds int64) (err error) {

	tmpProcessor, ok := p.delayQueueMap.Load(orgQueue)
	if !ok {
		actual, loaded := p.delayQueueMap.LoadOrStore(orgQueue, redisdelay.NewProcessor(p.client, orgQueue, p.opts.DelayInterval, p.Push))
		if !loaded {
			if processor, ok := actual.(queue.DelayProcessor); ok {
				processor.Start(p.closeChan)
			}
		}
		tmpProcessor = actual
	}
	return tmpProcessor.(queue.DelayProcessor).AppendMessage(msg, delaySeconds)
}
