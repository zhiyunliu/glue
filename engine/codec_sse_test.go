package engine

import (
	stdctx "context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gluectx "github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/golibs/xsse"
)

func TestDefaultResponseEncoder_ClearWriteDeadlineForSSE(t *testing.T) {
	resp := &writeDeadlineResponse{}
	ctx := &encoderContext{resp: resp}
	sse := &testSSE{events: []*xsse.Event{{Data: "hello"}}}

	err := DefaultResponseEncoder(ctx, sse)

	require.NoError(t, err)
	assert.True(t, resp.deadlineSet)
	assert.True(t, resp.deadline.IsZero())
}

func TestDefaultResponseEncoder_ClearWriteDeadlineForSSEv2(t *testing.T) {
	resp := &writeDeadlineResponse{}
	ctx := &encoderContext{resp: resp}
	sse := &testSSEv2{events: []*xsse.Event{{Data: "hello"}}}

	err := DefaultResponseEncoder(ctx, sse)

	require.ErrorIs(t, err, xsse.ErrChanIsEmpty)
	assert.True(t, resp.deadlineSet)
	assert.True(t, resp.deadline.IsZero())
	assert.Contains(t, string(resp.body), "data:hello\n\n")
}

func TestDefaultResponseEncoder_FlushHeaderBeforeFirstSSEEvent(t *testing.T) {
	resp := &writeDeadlineResponse{}
	ctx := &encoderContext{resp: resp}
	sse := &flushProbeSSE{resp: resp}

	err := DefaultResponseEncoder(ctx, sse)

	require.ErrorIs(t, err, xsse.ErrChanIsEmpty)
	assert.Equal(t, 1, sse.flushCountOnFirstEvent, "首个事件到来前应已刷出响应头")
}

// flushProbeSSE 记录第一次取事件时已发生的 Flush 次数，
// 用于验证响应头在事件循环开始前就已经发出。
type flushProbeSSE struct {
	resp                   *writeDeadlineResponse
	called                 bool
	flushCountOnFirstEvent int
}

func (s *flushProbeSSE) GetEventV2() (evt *xsse.Event, err error) {
	if !s.called {
		s.called = true
		s.flushCountOnFirstEvent = s.resp.flushCount
	}
	return nil, xsse.ErrChanIsEmpty
}

type testSSE struct {
	events []*xsse.Event
}

func (s *testSSE) GetEvent() (evt *xsse.Event, ok bool) {
	if len(s.events) == 0 {
		return nil, false
	}
	evt = s.events[0]
	s.events = s.events[1:]
	return evt, true
}

type testSSEv2 struct {
	events []*xsse.Event
}

func (s *testSSEv2) GetEventV2() (evt *xsse.Event, err error) {
	if len(s.events) == 0 {
		return nil, xsse.ErrChanIsEmpty
	}
	evt = s.events[0]
	s.events = s.events[1:]
	return evt, nil
}

type writeDeadlineResponse struct {
	headers     map[string]string
	body        []byte
	deadlineSet bool
	deadline    time.Time
	flushCount  int
}

func (r *writeDeadlineResponse) StatusCode(int) {}

func (r *writeDeadlineResponse) GetStatusCode() int { return 0 }

func (r *writeDeadlineResponse) GetHeader(key string) string {
	return r.headers[key]
}

func (r *writeDeadlineResponse) Header(key, val string) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}
	r.headers[key] = val
}

func (r *writeDeadlineResponse) Write(interface{}) error { return nil }

func (r *writeDeadlineResponse) WriteBytes(bytes []byte) error {
	r.body = append(r.body, bytes...)
	return nil
}

func (r *writeDeadlineResponse) ContentType() string { return "" }

func (r *writeDeadlineResponse) ResponseBytes() []byte { return r.body }

func (r *writeDeadlineResponse) Size() int { return len(r.body) }

func (r *writeDeadlineResponse) Redirect(int, string) {}

func (r *writeDeadlineResponse) Flush() error {
	r.flushCount++
	return nil
}

func (r *writeDeadlineResponse) SetWriteDeadline(deadline time.Time) error {
	r.deadlineSet = true
	r.deadline = deadline
	return nil
}

type encoderContext struct {
	resp gluectx.Response
}

func (c *encoderContext) GetImpl() interface{} { return nil }

func (c *encoderContext) ServerType() string { return "" }

func (c *encoderContext) ServerName() string { return "" }

func (c *encoderContext) ResetContext(stdctx.Context) {}

func (c *encoderContext) Context() stdctx.Context { return stdctx.Background() }

func (c *encoderContext) Header(string) string { return "" }

func (c *encoderContext) Request() gluectx.Request { return nil }

func (c *encoderContext) Bind(interface{}) error { return nil }

func (c *encoderContext) Response() gluectx.Response { return c.resp }

func (c *encoderContext) Meta() map[string]interface{} { return nil }

func (c *encoderContext) Log() log.Logger { return nil }

func (c *encoderContext) Close() {}
