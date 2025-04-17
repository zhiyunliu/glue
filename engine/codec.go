package engine

import (
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/engine"
	"github.com/zhiyunliu/golibs/xsse"
	"github.com/zhiyunliu/xbinding"
)

const (
	ContentTypeName = constants.ContentTypeName
)

// SupportPackageIsVersion1 These constants should not be referenced from any other code.
const SupportPackageIsVersion1 = true

type IoWriterWrapper = engine.IoWriterWrapper
type DataEncoder interface {
	Render(ctx context.Context) error
}

type ResponseEntity = engine.ResponseEntity

// DecodeRequestFunc is decode request func.
type DecodeRequestFunc func(context.Context, interface{}) error

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc func(context.Context, interface{}) error

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(context.Context, error)

// DefaultRequestDecoder decodes the request body to object.
func DefaultRequestDecoder(ctx context.Context, v interface{}) (err error) {

	method := ctx.Request().GetMethod()
	cttType := ctx.Request().GetHeader(ContentTypeName)
	mediaType, params, err := mime.ParseMediaType(cttType)
	if err != nil {
		return fmt.Errorf("DefaultRequestDecoder.ParseMediaType,%w,value=%s", err, cttType)
	}

	codec, err := xbinding.GetCodec(
		xbinding.WithMethod(method),
		xbinding.WithContentType(mediaType))
	if err != nil {
		return
	}

	//MethodGet
	if strings.EqualFold(string(MethodGet), method) &&
		strings.EqualFold(codec.ContentType(), xbinding.MIMEPOSTForm) {
		return codec.Bind(xbinding.MapReader(ctx.Request().Query().GetValues()), v)
	}

	//MIMEMultipartPOSTForm
	if strings.EqualFold(mediaType, xbinding.MIMEMultipartPOSTForm) {
		return codec.Bind(&xbinding.ReaderWrapper{
			Data: &xbinding.MultipartReqestInfo{
				Boundary: params["boundary"],
				Body:     ctx.Request().Body(),
			},
		}, v)
	}

	//normal
	return codec.Bind(&xbinding.ReaderWrapper{
		Data: ctx.Request().Body(),
	}, v)
}

// DefaultResponseEncoder encodes the object to the HTTP response.
func DefaultResponseEncoder(ctx context.Context, v interface{}) (err error) {
	if render, ok := v.(DataEncoder); ok {
		return render.Render(ctx)
	}
	resp := ctx.Response()

	if sseEntity, ok := v.(ServerSentEvents); ok {
		resp.Header(constants.ContentTypeName, xsse.ContentType)
		resp.Header(http.CanonicalHeaderKey("Connection"), "keep-alive")
		if cacheVal := resp.GetHeader(constants.ContentTypeCacheControl); cacheVal == "" {
			resp.Header(constants.ContentTypeCacheControl, constants.ContentTypeNoCache)
		}
		for {
			evt, ok := sseEntity.GetEvent()
			if !ok {
				break
			}
			err := xsse.Encode(IoWriterWrapper(resp.WriteBytes), evt)
			if err != nil {
				return err
			}
			resp.Flush()
		}
		return nil
	}

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
			return err
		}
		err = resp.WriteBytes(bytes)
		return err
	}

	var codec xbinding.Codec
	if _, ok := v.(string); ok {
		codec, _ = xbinding.GetCodec(xbinding.WithContentType("text"))
	} else {
		codec, _ = CodecForRequest(ctx, "Accept")
	}

	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	resp.Header(ContentTypeName, codec.ContentType())
	err = resp.WriteBytes(data)
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
	resp.Header(ContentTypeName, codec.ContentType())
	resp.Status(se.Code)
	resp.WriteBytes(body)
}

// CodecForRequest get encoding.Codec via http.Request
func CodecForRequest(ctx context.Context, headerName string) (xbinding.Codec, bool) {
	headVal := ctx.Request().GetHeader(headerName)
	codec, err := xbinding.GetCodec(xbinding.WithContentType(headVal))
	if err == nil && codec != nil {
		return codec, true
	}
	codec, _ = xbinding.GetCodec(xbinding.WithContentType("json"))
	return codec, false
}
