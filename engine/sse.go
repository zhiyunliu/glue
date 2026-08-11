package engine

import (
	"net/http"
	"time"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/golibs/xsse"
)

// Deprecated: use xsse.ServerSentEvents in the future.
type ServerSentEvents = xsse.ServerSentEvents
type ServerSentEventv2 = xsse.ServerSentEventv2

type nextSSEEvent func() (xsse.SSEEvent, bool, error)

func processSSEStream(ctx context.Context, v any) (ok bool, err error) {
	ok, err = processSSEv2(ctx, v)
	if ok {
		return
	}
	return processSSEv1(ctx, v)
}

func processSSEv1(ctx context.Context, v any) (ok bool, err error) {
	sseEntity, ok := v.(ServerSentEvents)
	if !ok {
		return
	}
	err = processSSEEvents(ctx.Response(), func() (xsse.SSEEvent, bool, error) {
		evt, evtok := sseEntity.GetEvent()
		return evt, evtok, nil
	})
	return
}

func processSSEv2(ctx context.Context, v any) (ok bool, err error) {
	sseEntity, ok := v.(ServerSentEventv2)
	if !ok {
		return
	}
	err = processSSEEvents(ctx.Response(), func() (evt xsse.SSEEvent, ok bool, err error) {
		evt, err = sseEntity.GetEventV2()
		if err != nil {
			return
		}
		return evt, true, nil
	})
	return
}

func processSSEEvents(resp context.Response, nextEvent nextSSEEvent) (err error) {
	prepareSSEResponse(resp)
	// 先把响应头发出去，避免首个事件迟迟不来时客户端触发 response header 超时。
	flushSSEResponse(resp)
	defer flushSSEResponse(resp)

	for {
		evt, ok, nextErr := nextEvent()
		if nextErr != nil {
			return nextErr
		}
		if !ok {
			return nil
		}
		if err = xsse.Encode(IoWriterWrapper(resp.WriteBytes), evt); err != nil {
			return err
		}
		resp.Flush()
	}
}

func prepareSSEResponse(resp context.Response) {
	if deadlineSetter, ok := resp.(context.WriteDeadlineSetter); ok {
		_ = deadlineSetter.SetWriteDeadline(time.Time{})
	}
	resp.Header(constants.ContentTypeName, xsse.ContentType)
	resp.Header(http.CanonicalHeaderKey("Connection"), "keep-alive")
	if cacheVal := resp.GetHeader(constants.ContentTypeCacheControl); cacheVal == "" {
		resp.Header(constants.ContentTypeCacheControl, constants.ContentTypeNoCache)
	}
}

func flushSSEResponse(resp context.Response) {
	_ = resp.WriteBytes([]byte{})
	resp.Flush()
}
