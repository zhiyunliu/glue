package queues

//IQueue 消息队列
type IQueue interface {
	Send(key string, value interface{}) error
	Pop(key string) (string, error)
	Count(key string) (int64, error)
}

//IComponentQueue Component Queue
type IComponentQueue interface {
	GetQueue(name string) (q IQueue)
}
