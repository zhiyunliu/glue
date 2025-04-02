package http

import (
	"net/http"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/xhttp"
	"github.com/zhiyunliu/golibs/bytesconv"
)

var (
	_ xhttp.Body            = (*errorBody)(nil)
	_ engine.ResponseEntity = (*errorBody)(nil)
)

func newBodyByError(err error) xhttp.Body {
	return &errorBody{
		err: err,
		header: map[string]string{
			constants.ContentTypeName: constants.ContentTypeTextPlain,
		},
	}
}

type errorBody struct {
	err    error
	header map[string]string
}

func (b *errorBody) GetStatus() int32 {
	return http.StatusInternalServerError
}
func (b *errorBody) GetHeader() map[string]string {
	return b.header
}
func (b *errorBody) GetResult() []byte {
	return bytesconv.StringToBytes(b.err.Error())
}

func (b *errorBody) StatusCode() int {
	return int(b.GetStatus())
}
func (b *errorBody) Header() map[string]string {
	return b.GetHeader()
}
func (b *errorBody) Body() (bytes []byte, err error) {
	return b.GetResult(), nil
}
