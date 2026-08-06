package gin

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gingonic "github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

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
