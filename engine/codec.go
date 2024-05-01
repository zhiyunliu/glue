package engine

import (
	"io"
	"net/http"
	"strings"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/context"

	"github.com/zhiyunliu/glue/encoding"
	"github.com/zhiyunliu/glue/encoding/text"
	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/golibs/httputil"
)

const (
	ContentTypeName = constants.ContentTypeName
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
func DefaultRequestDecoder(ctx context.Context, v interface{}) (err error) {
	var data []byte
	if strings.EqualFold(string(MethodGet), ctx.Request().GetMethod()) {
		data = []byte(ctx.Request().Query().String())
	} else {
		data, err = io.ReadAll(ctx.Request().Body())
		if err != nil {
			return errors.BadRequest("CODEC", err.Error())
		}
	}

	codec, ok := CodecForRequest(ctx, ContentTypeName, ctx.Request().GetMethod())
	if !ok {
		return errors.BadRequest("CODEC", ctx.Request().GetHeader(ContentTypeName))
	}
	if err = codec.Unmarshal(data, v); err != nil {
		return errors.BadRequest("CODEC", err.Error())
	}
	return nil
}

// DefaultResponseEncoder encodes the object to the HTTP response.
func DefaultResponseEncoder(ctx context.Context, v interface{}) (err error) {
	if render, ok := v.(DataEncoder); ok {
		return render.Render(ctx)
	}

	//判定对象是否实现了响应体接口
	if entity, ok := v.(ResponseEntity); ok {
		resp := ctx.Response()

		resp.Status(entity.StatusCode())
		for k, v := range entity.Header() {
			resp.Header(k, v)
		}
		bytes, err := entity.Body()
		if err != nil {
			return err
		}
		err = resp.WriteBytes(bytes)
		return err
	}

	var codec encoding.Codec
	if _, ok := v.(string); ok {
		codec = encoding.GetCodec(text.Name)
	} else {
		codec, _ = CodecForRequest(ctx, "Accept", "")
	}

	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	ctx.Response().Header(ContentTypeName, httputil.ContentType(codec.Name()))
	err = ctx.Response().WriteBytes(data)
	return err
}

// DefaultErrorEncoder encodes the error to the HTTP response.
func DefaultErrorEncoder(ctx context.Context, err error) {
	se := errors.FromError(err)
	codec, _ := CodecForRequest(ctx, "Accept", "")
	body, err := codec.Marshal(se)
	if err != nil {
		ctx.Response().Status(500)
		return
	}
	ctx.Response().Header(ContentTypeName, httputil.ContentType(codec.Name()))
	ctx.Response().Status(se.Code)
	ctx.Response().WriteBytes(body)
}

// CodecForRequest get encoding.Codec via http.Request
func CodecForRequest(ctx context.Context, headerName string, method string) (encoding.Codec, bool) {
	headVal := ctx.Request().GetHeader(headerName)
	if method != "" && strings.EqualFold(method, http.MethodGet) {
		codec := encoding.GetCodec("form")
		return codec, true
	}

	codec := encoding.GetCodec(httputil.ContentSubtype(headVal))
	if codec != nil {
		return codec, true
	}
	return encoding.GetCodec("json"), false
}
