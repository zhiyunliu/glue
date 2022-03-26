package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xtypes"
	"github.com/zhiyunliu/velocity/log"

	"github.com/gin-gonic/gin"
	vctx "github.com/zhiyunliu/velocity/context"
)

const XRequestId = "x-request-id"

type GinContext struct {
	srvType string
	opts    *options
	Gctx    *gin.Context
	greq    *ginRequest
	gresp   *gresponse
	logger  log.Logger
}

func (ctx *GinContext) reset(gctx *gin.Context) {
	ctx.Gctx = gctx
	ctx.greq = nil
}

func (ctx *GinContext) ServerType() string {
	return ctx.srvType
}

func (ctx *GinContext) Context() context.Context {
	return ctx.Gctx.Request.Context()
}

func (ctx *GinContext) ResetContext(nctx context.Context) {
	req := ctx.Gctx.Request.WithContext(nctx)
	ctx.Gctx.Request = req
}

func (ctx *GinContext) Header(key string) string {
	return ctx.Gctx.GetHeader(key)
}

func (ctx *GinContext) Request() vctx.Request {
	if ctx.greq == nil {
		ctx.greq = &ginRequest{gctx: ctx.Gctx}
	}
	return ctx.greq
}
func (ctx *GinContext) Response() vctx.Response {
	if ctx.gresp == nil {
		ctx.gresp = &gresponse{gctx: ctx.Gctx, vctx: ctx}
	}
	return ctx.gresp
}
func (ctx *GinContext) Log() log.Logger {
	if ctx.logger == nil {
		logger, ok := log.FromContext(ctx.Context())
		if !ok {
			xreqId := ctx.Gctx.GetHeader(XRequestId)
			if xreqId == "" {
				xreqId = session.Create()
				ctx.Gctx.Header(XRequestId, xreqId)
			}
			logger = log.New(log.WithName("gin"), log.WithSid(xreqId))
			ctx.ResetContext(log.WithContext(ctx.Context(), logger))
		}
		ctx.logger = logger
	}
	return ctx.logger
}
func (ctx *GinContext) Close() {

	if ctx.greq.gpath != nil {
		ctx.greq.gpath.params = nil
		ctx.greq.gpath.gctx = nil
		ctx.greq.gpath = nil
	}

	if ctx.greq.gquery != nil {
		ctx.greq.gquery.params = nil
		ctx.greq.gquery.gctx = nil
		ctx.greq.gquery = nil
	}

	if ctx.greq.gbody != nil {
		ctx.greq.gbody.bodyBytes = nil
		ctx.greq.gbody.gctx = nil
		ctx.greq.gbody = nil
	}

	if ctx.greq.gpath != nil {
		ctx.greq.gpath.params = nil
		ctx.greq.gpath.gctx = nil
		ctx.greq.gpath = nil
	}

	if ctx.gresp != nil {
		ctx.gresp.gctx = nil
		ctx.gresp = nil
	}

	ctx.greq.gctx = nil
	ctx.Gctx = nil

	ctx.logger.Close()
	ctx.logger = nil
}

func (ctx *GinContext) GetImpl() interface{} {
	return ctx.Gctx
}

//--------------------------------

type ginRequest struct {
	gctx   *gin.Context
	gpath  *gpath
	gquery *gquery
	gbody  *gbody
}

func (r *ginRequest) GetMethod() string {
	return r.gctx.Request.Method
}

func (r *ginRequest) GetClientIP() string {
	return r.gctx.ClientIP()
}

// func (r *ginRequest) Header() vctx.Header {
// 	return r.gctx.GetHeader(key)
// }

func (r *ginRequest) GetHeader(key string) string {
	return r.gctx.GetHeader(key)
}

func (r *ginRequest) SetHeader(key, val string) {
	r.gctx.Header(key, val)
}

func (r *ginRequest) Path() vctx.Path {
	if r.gpath == nil {
		r.gpath = &gpath{gctx: r.gctx}
	}
	return r.gpath
}

func (r *ginRequest) Query() vctx.Query {
	if r.gquery == nil {
		r.gquery = &gquery{gctx: r.gctx}
	}
	return r.gquery
}
func (r *ginRequest) Body() vctx.Body {
	if r.gbody == nil {
		r.gbody = &gbody{gctx: r.gctx}
	}
	return r.gbody
}

//-path-------------------------

type gpath struct {
	gctx   *gin.Context
	params xtypes.SMap
}

func (p *gpath) GetURL() *url.URL {
	return p.gctx.Request.URL
}

func (p *gpath) FullPath() string {
	return p.gctx.FullPath()
}
func (p *gpath) Params() xtypes.SMap {
	if p.params == nil {
		p.params = xtypes.SMap{}
		tps := p.gctx.Params
		for i := range tps {
			p.params[tps[i].Key] = tps[i].Value
		}
	}
	return p.params
}

//-gquery---------------------------------

type gquery struct {
	gctx   *gin.Context
	params xtypes.SMap
}

func (q *gquery) Get(name string) string {
	return q.gctx.Query(name)
}
func (q *gquery) SMap() xtypes.SMap {
	if q.params == nil {
		vals := q.gctx.Request.URL.Query()
		q.params = make(xtypes.SMap)
		for k := range vals {
			q.params[k] = vals.Get(k)
		}
	}
	return q.params
}
func (q *gquery) Scan(obj interface{}) error {
	return q.gctx.BindQuery(obj)
}

func (q *gquery) String() string {
	return q.gctx.Request.URL.RawQuery
}

//-gbody---------------------------------
type gbody struct {
	gctx      *gin.Context
	hasRead   bool
	bodyBytes []byte
}

func (q *gbody) Scan(obj interface{}) error {
	return q.gctx.Bind(obj)
}

func (q *gbody) Read(p []byte) (n int, err error) {
	err = q.loadBody()
	if err != nil {
		return
	}
	return bytes.NewReader(q.bodyBytes).Read(p)
}

func (q *gbody) Len() int {
	err := q.loadBody()
	if err != nil {
		return 0
	}
	return len(q.bodyBytes)
}

func (q *gbody) Bytes() []byte {
	err := q.loadBody()
	if err != nil {
		return nil
	}
	return q.bodyBytes
}

func (q *gbody) loadBody() (err error) {
	if len(q.bodyBytes) == 0 && !q.hasRead {
		q.hasRead = true
		q.bodyBytes, err = ioutil.ReadAll(q.gctx.Request.Body)
		if err != nil {
			return err
		}
		q.gctx.Request.Body.Close()
	}
	return nil
}

//gresponse --------------------------------
type gresponse struct {
	vctx      *GinContext
	gctx      *gin.Context
	hasWrited bool
}

func (q *gresponse) Status(statusCode int) {
	q.gctx.Writer.WriteHeader(statusCode)
}

func (q *gresponse) Header(key, val string) {
	q.gctx.Writer.Header().Set(key, val)
}

func (q *gresponse) ContextType(val string) {
	q.gctx.Writer.Header().Set("content-type", val)
}

func (q *gresponse) Write(obj interface{}) error {
	if q.hasWrited {
		panic(fmt.Errorf("%s：有多种方式写入响应", q.gctx.FullPath()))
	}
	q.hasWrited = true
	if werr, ok := obj.(error); ok {
		q.vctx.opts.ErrorEncoder(q.vctx, werr)
		return nil
	}
	return q.vctx.opts.ResponseEncoder(q.vctx, obj)
}

func (q *gresponse) WriteBytes(bytes []byte) error {
	_, err := q.gctx.Writer.Write(bytes)
	return err
}
