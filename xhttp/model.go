package xhttp

import (
	"net/http"

	"github.com/zhiyunliu/glue/constants"
)

var emptyBodyObj = &emptyBody{
	header: map[string]string{
		constants.ContentTypeName: constants.ContentTypeTextPlain,
	},
}

// NewEmptyBody returns an empty body object.
func NewEmptyBody() Body {
	return emptyBodyObj
}

type emptyBody struct {
	header map[string]string
}

func (b *emptyBody) GetStatus() int32 {
	return http.StatusOK
}
func (b *emptyBody) GetHeader() map[string]string {
	return b.header
}
func (b *emptyBody) GetResult() []byte {
	return []byte{}
}
