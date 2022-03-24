package server

import (
	"io"

	"github.com/zhiyunliu/velocity/context"

	"github.com/zhiyunliu/golibs/httputil"
	"github.com/zhiyunliu/velocity/encoding"
	"github.com/zhiyunliu/velocity/errors"
)

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

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
	codec, _ := CodecForRequest(ctx, "Accept")
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
	se := errors.FromError(err)
	codec, _ := CodecForRequest(ctx, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		ctx.Response().Status(500)
		return
	}
	ctx.Response().Header("Content-Type", httputil.ContentType(codec.Name()))
	ctx.Response().Status(se.Code)
	ctx.Response().WriteBytes(body)
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
