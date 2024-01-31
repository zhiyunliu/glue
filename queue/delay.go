package queue

//DelayCallback 延迟消息处理回调
type DelayCallback func(key string, msgList ...Message) error

// DelayProcessor 延迟消息队列处理器
type DelayProcessor interface {
	Start(done chan struct{})
	AppendMessage(msg Message, delaySeconds int64) error
}
