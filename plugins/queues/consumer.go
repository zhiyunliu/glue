package queues

type ConsumerFunc func(Messager) error
