package engine

import (
	"io"
	"net/http"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/context"

	"github.com/zhiyunliu/glue/encoding"
	"github.com/zhiyunliu/glue/encoding/text"
	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/httputil"
)

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

type DataEncoder interface {
	Render(ctx context.Context) error
}

type ResponseEntity interface {
	StatusCode() int
	Header() map[string]string
	Body() (bytes []byte, err error)
}

// DecodeRequestFunc is decode request func.
type DecodeRequestFunc func(context.Context, interface{}) error

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc func(context.Context, interface{}) error

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(context.Context, error)

// DefaultRequestDecoder decodes the request body to object.
func DefaultRequestDecoder(ctx context.Context, v interface{}) error {

	codec, ok := CodecForRequest(ctx, "Content-Type")
	if !ok {
		return errors.BadRequest("CODEC", ctx.Request().GetHeader("Content-Type"))
	}
	data, err := io.ReadAll(ctx.Request().Body())
	if err != nil {
		return errors.BadRequest("CODEC", err.Error())
	}
	if err = codec.Unmarshal(data, v); err != nil {
		return errors.BadRequest("CODEC", err.Error())
	}
	return nil
}

// DefaultResponseEncoder encodes the object to the HTTP response.
func DefaultResponseEncoder(ctx context.Context, v interface{}) error {
	if render, ok := v.(DataEncoder); ok {
		return render.Render(ctx)
	}

	resp := ctx.Response()

	//判定对象是否实现了响应体接口
	if entity, ok := v.(ResponseEntity); ok {
		resp.Status(entity.StatusCode())
		header := entity.Header()
		if len(header) > 0 {
			for k, v := range header {
				resp.Header(k, v)
			}
		}
		bytes, err := entity.Body()
		if err != nil {
			resp.Status(http.StatusInternalServerError)
			resp.WriteBytes(bytesconv.StringToBytes(err.Error()))
			return err
		}
		return resp.WriteBytes(bytes)
	}

	var codec encoding.Codec
	if _, ok := v.(string); ok {
		codec = encoding.GetCodec(text.Name)
	} else {
		codec, _ = CodecForRequest(ctx, "Accept")
	}

	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	ctx.Response().Header("Content-Type", httputil.ContentType(codec.Name()))
	err = ctx.Response().WriteBytes(data)

	return err
}

// DefaultErrorEncoder encodes the error to the HTTP response.
func DefaultErrorEncoder(ctx context.Context, err error) {
	if render, ok := err.(DataEncoder); ok {
		render.Render(ctx)
		return
	}
	resp := ctx.Response()

	//判定对象是否实现了响应体接口
	if entity, ok := err.(ResponseEntity); ok {
		resp.Status(entity.StatusCode())
		header := entity.Header()
		if len(header) > 0 {
			for k, v := range header {
				resp.Header(k, v)
			}
		}
		bytes, err := entity.Body()
		if err != nil {
			resp.Status(http.StatusInternalServerError)
			resp.WriteBytes(bytesconv.StringToBytes(err.Error()))
			return
		}
		resp.WriteBytes(bytes)
		return
	}

	se := errors.FromError(err)
	codec, _ := CodecForRequest(ctx, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		resp.Status(http.StatusInternalServerError)
		resp.WriteBytes(bytesconv.StringToBytes(err.Error()))
		return
	}

	resp.Header(constants.ContentTypeName, httputil.ContentType(codec.Name()))
	resp.Status(se.Code)
	resp.WriteBytes(body)
}

// CodecForRequest get encoding.Codec via http.Request
func CodecForRequest(ctx context.Context, name string) (encoding.Codec, bool) {
	accept := ctx.Request().GetHeader(name)
	codec := encoding.GetCodec(httputil.ContentSubtype(accept))
	if codec != nil {
		return codec, true
	}
	return encoding.GetCodec("json"), false
}
