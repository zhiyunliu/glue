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
