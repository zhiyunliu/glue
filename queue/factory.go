package queue

import (
	"errors"

	"github.com/zhiyunliu/gel/config"
)

var Nil error = errors.New("Queue Nil")

//queue 对输入KEY进行封装处理
type queue struct {
	q IMQP
}

func newQueue(setting config.Config) (IQueue, error) {
	var err error
	q := &queue{}
	q.q, err = NewMQP(setting)
	return q, err
}

//Send 发送消息
func (q *queue) Send(key string, value Message) error {

	return q.q.Push(key, value)
}

//Pop 从队列中获取一个消息
func (q *queue) Pop(key string) (string, error) {
	return q.q.Pop(key)
}

//Count 队列中消息个数
func (q *queue) Count(key string) (int64, error) {
	return q.q.Count(key)
}

func (q *queue) Close() error {
	return q.q.Close()
}
