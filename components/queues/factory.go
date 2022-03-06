package queues

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/zhiyunliu/velocity/components/queues/impls"
	"github.com/zhiyunliu/velocity/config"
)

//queue 对输入KEY进行封装处理
type queue struct {
	q impls.IMQP
}

func newQueue(setting *config.Setting) (IQueue, error) {
	var err error
	q := &queue{}
	q.q, err = impls.NewMQP(setting)
	return q, err
}

//Send 发送消息
func (q *queue) Send(key string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		err = fmt.Errorf("发送消息队列:%s失败:%+v", key, err)
		return err
	}

	return q.q.Push(key, string(bytes))
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
