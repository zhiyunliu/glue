package engine

import (
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/golibs/engine"
	"github.com/zhiyunliu/xbinding"
)

// ResponseWriter ...
type ResponseWriter = engine.ResponseWriter

func processEntity(ctx context.Context, v any) (ok bool, err error) {
	resp := ctx.Response()
	//判定对象是否实现了响应体接口
	entity, ok := v.(ResponseEntity)
	if !ok {
		return false, nil
	}
	resp.StatusCode(entity.StatusCode())
	header := entity.Header()
	if len(header) > 0 {
		for k, v := range header {
			resp.Header(k, v)
		}
	}
	bytes, err := entity.Body()
	if err != nil {
		return
	}
	err = resp.WriteBytes(bytes)
	return

}

func processDefault(ctx context.Context, v any) (err error) {
	resp := ctx.Response()

	var codec xbinding.Codec
	if _, ok := v.(string); ok {
		codec, _ = xbinding.GetCodec(xbinding.WithContentType("text"))
	} else {
		codec, _ = CodecForRequest(ctx, "Accept")
	}

	data, err := codec.Marshal(v)
	if err != nil {
		return
	}
	resp.Header(ContentTypeName, codec.ContentType())
	err = resp.WriteBytes(data)
	return err
}
