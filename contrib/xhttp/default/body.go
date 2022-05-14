package grpc

import (
	"net/http"

	"github.com/zhiyunliu/gel/constants"
	"github.com/zhiyunliu/gel/xhttp"
	"github.com/zhiyunliu/golibs/bytesconv"
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
