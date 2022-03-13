package mqc

import (
	"net/url"

	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/queue"
)

//Request 处理任务请求
type Request struct {
	task *Task
	queue.IMQCMessage
	method string
	form   context.Body
	header context.Header
}

//NewRequest 构建任务请求
func NewRequest(task *Task, m queue.IMQCMessage) (r *Request, err error) {

	r = &Request{
		IMQCMessage: m,
		task:        task,
		method:      "GET",
	}

	//将消息原串转换为map
	message := m.GetMessage()

	r.header = message.Header()
	r.form = message.Body()

	return r, nil
}

//GetName 获取任务名称
func (m *Request) GetName() string {
	return m.task.Queue
}

//GetService 服务名
func (m *Request) GetService() string {
	return m.task.GetService()
}

//GetMethod 方法名
func (m *Request) GetMethod() string {
	return m.method
}

//GetForm 输入参数
func (m *Request) GetBody() context.Body {
	return m.form
}

func (m *Request) GetForm() url.Values {
	return url.Values{}
}
func (m *Request) GetHeader() map[string]string {
	return nil
}
func (m *Request) GetRemoteAddr() string {
	return ""
}
