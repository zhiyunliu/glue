package gin

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gingonic "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	vctx "github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/golibs/xsse"
)

// blockingSSE 模拟“连接已建立但长时间没有事件”的上游，
// 例如只等待客户端断开的保活型 SSE 接口。
type blockingSSE struct {
	release <-chan struct{}
}

func (s *blockingSSE) GetEventV2() (*xsse.Event, error) {
	<-s.release
	return nil, xsse.ErrChanIsEmpty
}

func TestGinSSEResponseHeaderArrivesBeforeFirstEvent(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)

	release := make(chan struct{})
	ginEngine := gingonic.New()
	adapter := NewGinEngine(ginEngine, engine.WithSrvType("api"))
	adapter.Handle(http.MethodGet, "/sse", func(ctx vctx.Context) {
		adapter.Write(ctx, &blockingSSE{release: release})
	})

	server := httptest.NewServer(adapter.GetImpl().(http.Handler))
	t.Cleanup(server.Close)
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})

	// 客户端的 ResponseHeaderTimeout 远小于上游首个事件的到达时间，
	// 修复前会因为响应头迟迟不发出而超时。
	client := &http.Client{
		Transport: &http.Transport{ResponseHeaderTimeout: 500 * time.Millisecond},
	}
	response, err := client.Get(server.URL + "/sse")
	require.NoError(t, err, "响应头应在首个事件之前到达")
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, xsse.ContentType, response.Header.Get("Content-Type"))

	close(release)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Empty(t, string(body))
}

func TestGinResponse_SetWriteDeadlineAllowsSSEAfterServerWriteTimeout(t *testing.T) {
	gingonic.SetMode(gingonic.TestMode)

	router := gingonic.New()
	router.GET("/sse", func(gctx *gingonic.Context) {
		resp := &ginResponse{gctx: gctx}
		require.NoError(t, resp.SetWriteDeadline(time.Time{}))

		gctx.Header("Content-Type", "text/event-stream")
		time.Sleep(80 * time.Millisecond)
		_, err := gctx.Writer.Write([]byte("data:ok\n\n"))
		require.NoError(t, err)
		gctx.Writer.Flush()
	})

	server := httptest.NewUnstartedServer(router)
	server.Config.WriteTimeout = 30 * time.Millisecond
	server.Start()
	t.Cleanup(server.Close)

	response, err := server.Client().Get(server.URL + "/sse")
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, "data:ok\n\n", string(body))
}
