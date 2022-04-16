package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/zhiyunliu/gel/constants"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xtypes"

	"github.com/gin-gonic/gin"
	vctx "github.com/zhiyunliu/gel/context"
)

type GinContext struct {
	opts   *options
	Gctx   *gin.Context
	greq   *ginRequest
	gresp  *ginResponse
	logger log.Logger
}

func newGinContext(opts *options) *GinContext {
	return &GinContext{
		opts: opts,
		greq: &ginRequest{
			closed: true,
			gpath:  &gpath{closed: true},
			gquery: &gquery{closed: true},
			gbody:  &gbody{closed: true},
		},
		gresp: &ginResponse{closed: true},
	}
}

func (ctx *GinContext) reset(gctx *gin.Context) {
	ctx.Gctx = gctx
}

func (ctx *GinContext) ServerType() string {
	return ctx.opts.SrvType
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
	if ctx.greq.closed {
		ctx.greq.closed = false
		ctx.greq.gctx = ctx.Gctx
	}
	return ctx.greq
}
func (ctx *GinContext) Response() vctx.Response {
	if ctx.gresp.closed {
		ctx.gresp.closed = false
		ctx.gresp.gctx = ctx.Gctx
		ctx.gresp.vctx = ctx
	}
	return ctx.gresp
}
func (ctx *GinContext) Log() log.Logger {
	if ctx.logger == nil {
		logger, ok := log.FromContext(ctx.Context())
		if !ok {
			xreqId := ctx.Gctx.GetHeader(constants.HeaderRequestId)
			if xreqId == "" {
				xreqId = session.Create()
				ctx.Gctx.Header(constants.HeaderRequestId, xreqId)
			}
			logger = log.New(log.WithName("gin"), log.WithSid(xreqId))
			ctx.ResetContext(log.WithContext(ctx.Context(), logger))
		}
		ctx.logger = logger
	}
	return ctx.logger
}
func (ctx *GinContext) Close() {

	ctx.greq.Close()
	ctx.gresp.Close()
	ctx.Gctx = nil

	if ctx.logger != nil {
		ctx.logger.Close()
	}
	ctx.logger = nil
}

func (ctx *GinContext) GetImpl() interface{} {
	return ctx.Gctx
}

//--------------------------------

type ginRequest struct {
	gctx    *gin.Context
	gheader map[string]string
	gpath   *gpath
	gquery  *gquery
	gbody   *gbody
	closed  bool
}

func (r *ginRequest) GetMethod() string {
	return r.gctx.Request.Method
}

func (r *ginRequest) GetClientIP() string {
	return r.gctx.ClientIP()
}

func (r *ginRequest) Header() map[string]string {
	if r.gheader == nil {
		r.gheader = map[string]string{}
		gheader := r.gctx.Request.Header
		for k, v := range gheader {
			r.gheader[k] = strings.Join(v, ",")
		}
	}

	return r.gheader
}

func (r *ginRequest) GetHeader(key string) string {
	return r.gctx.GetHeader(key)
}

func (r *ginRequest) SetHeader(key, val string) {
	r.gctx.Header(key, val)
}

func (r *ginRequest) Path() vctx.Path {
	if r.gpath.closed {
		r.gpath.gctx = r.gctx
		r.gpath.closed = false
	}
	return r.gpath
}

func (r *ginRequest) Query() vctx.Query {
	if r.gquery.closed {
		r.gquery.gctx = r.gctx
		r.gquery.closed = false
	}
	return r.gquery
}
func (r *ginRequest) Body() vctx.Body {
	if r.gbody.closed {
		r.gbody.gctx = r.gctx
		r.gbody.closed = false
	}
	return r.gbody
}
func (q *ginRequest) Close() {
	q.closed = true
	q.gctx = nil
	q.gheader = nil
	q.gpath.Close()
	q.gquery.Close()
	q.gbody.Close()
}

//-path-------------------------

type gpath struct {
	gctx   *gin.Context
	params xtypes.SMap
	closed bool
}

func (p *gpath) GetURL() *url.URL {
	return p.gctx.Request.URL
}

func (p *gpath) FullPath() string {
	if p.gctx == nil {
		fmt.Println("xxx")
	}
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
func (q *gpath) Close() {
	q.gctx = nil
	q.params = nil
	q.closed = true
}

//-gquery---------------------------------

type gquery struct {
	gctx   *gin.Context
	params xtypes.SMap
	closed bool
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

func (q *gquery) Close() {
	q.gctx = nil
	q.params = nil
	q.closed = true
}

//-gbody---------------------------------
type gbody struct {
	gctx      *gin.Context
	hasRead   bool
	bodyBytes []byte
	reader    *bytes.Reader
	closed    bool
}

func (q *gbody) Scan(obj interface{}) error {
	return q.gctx.Bind(obj)
}

func (q *gbody) Read(p []byte) (n int, err error) {
	err = q.loadBody()
	if err != nil {
		return
	}
	return q.reader.Read(p)
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
		q.reader = bytes.NewReader(q.bodyBytes)
		q.gctx.Request.Body.Close()
	}
	return nil
}
func (q *gbody) Close() {
	q.bodyBytes = nil
	q.reader = nil
	q.gctx = nil
	q.closed = true
	q.hasRead = false
}

//gresponse --------------------------------
type ginResponse struct {
	vctx      *GinContext
	gctx      *gin.Context
	hasWrited bool
	closed    bool
}

func (q *ginResponse) Status(statusCode int) {
	q.gctx.Writer.WriteHeader(statusCode)
}

func (q *ginResponse) Header(key, val string) {
	q.gctx.Writer.Header().Set(key, val)
}

func (q *ginResponse) ContextType(val string) {
	q.gctx.Writer.Header().Set(constants.ContentTypeName, val)
}

func (q *ginResponse) Write(obj interface{}) error {
	if q.hasWrited {
		panic(fmt.Errorf("%s:有多种方式写入响应", q.gctx.FullPath()))
	}
	q.hasWrited = true
	if werr, ok := obj.(error); ok {
		q.vctx.opts.ErrorEncoder(q.vctx, werr)
		return nil
	}
	return q.vctx.opts.ResponseEncoder(q.vctx, obj)
}

func (q *ginResponse) WriteBytes(bytes []byte) error {
	_, err := q.gctx.Writer.Write(bytes)
	return err
}

func (q *ginResponse) Close() {
	q.vctx = nil
	q.gctx = nil
	q.hasWrited = false
	q.closed = true
}
