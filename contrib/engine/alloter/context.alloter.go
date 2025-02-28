package alloter

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	"github.com/zhiyunliu/alloter"
	"github.com/zhiyunliu/glue/constants"
	vctx "github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/log"
	gluesid "github.com/zhiyunliu/glue/session"

	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xtypes"
)

type AlloterContext struct {
	Actx   *alloter.Context
	opts   *engine.Options
	meta   map[string]interface{}
	areq   *alloterRequest
	aresp  *alloterResponse
	logger log.Logger
}

func newAlloterContext(opts *engine.Options) *AlloterContext {
	return &AlloterContext{
		opts: opts,
		meta: make(map[string]interface{}),
		areq: &alloterRequest{
			closed: true,
			apath:  &apath{closed: true},
			aquery: &aquery{closed: true},
			abody:  &abody{closed: true},
		},
		aresp: &alloterResponse{
			closed: true,
		},
	}
}

func (ctx *AlloterContext) LogOptions() *log.Options {
	return ctx.opts.LogOpts
}

func (ctx *AlloterContext) reset(gctx *alloter.Context) {
	ctx.Actx = gctx
}

func (ctx *AlloterContext) ServerType() string {
	return ctx.opts.SrvType
}

func (ctx *AlloterContext) ServerName() string {
	return ctx.opts.SrvName
}

func (ctx *AlloterContext) Meta() map[string]interface{} {
	return ctx.meta
}
func (ctx *AlloterContext) Context() context.Context {
	return ctx.Actx.Request.Context()
}

func (ctx *AlloterContext) ResetContext(nctx context.Context) {
	ctx.Actx.Request.WithContext(nctx)
}

func (ctx *AlloterContext) Header(key string) string {
	return ctx.Actx.GetHeader(key)
}

func (ctx *AlloterContext) Bind(obj interface{}) error {
	val := reflect.TypeOf(obj)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("Bind只接收Ptr类型的数据,目前是:%s", val.Kind())
	}

	err := ctx.Request().Body().Scan(obj)
	if err != nil {
		return err
	}
	if chr, ok := obj.(engine.IChecker); ok {
		return chr.Check()
	}
	return nil
}

func (ctx *AlloterContext) Request() vctx.Request {
	if ctx.areq.closed {
		ctx.areq.closed = false
		ctx.areq.actx = ctx.Actx
		ctx.areq.vctx = ctx
		ctx.areq.reqUrl = ctx.Actx.Request.GetURL()
	}
	return ctx.areq
}
func (ctx *AlloterContext) Response() vctx.Response {
	if ctx.aresp.closed {
		ctx.aresp.closed = false
		ctx.aresp.actx = ctx.Actx
		ctx.aresp.vctx = ctx
	}
	return ctx.aresp
}
func (ctx *AlloterContext) Log() log.Logger {
	if ctx.logger == nil {
		orgCtx := ctx.Context()
		logger, ok := log.FromContext(orgCtx)
		if !ok {
			xreqId := ctx.Actx.GetHeader(constants.HeaderRequestId)
			if xreqId == "" {
				xreqId = session.Create()
				ctx.Actx.Header(constants.HeaderRequestId, xreqId)
			}
			orgCtx = gluesid.WithContext(orgCtx, xreqId)

			logger = log.New(orgCtx, log.WithName("alloter"),
				log.WithSid(xreqId),
				log.WithSrvType(ctx.opts.SrvType),
				log.WithField("src_name", ctx.Request().GetHeader(constants.HeaderSourceName)),
				log.WithField("src_ip", ctx.Request().GetHeader(constants.HeaderSourceIp)),
				log.WithField("cip", ctx.Request().GetClientIP()),
				log.WithField("uid", ctx.Actx.GetHeader(constants.AUTH_USER_ID)),
			)

			ctx.ResetContext(log.WithContext(orgCtx, logger))
		}
		ctx.logger = logger
	}
	return ctx.logger
}
func (ctx *AlloterContext) Close() {
	ctx.areq.Close()
	ctx.aresp.Close()
	ctx.Actx = nil
	ctx.meta = make(map[string]interface{})
	if ctx.logger != nil {
		ctx.logger.Close()
	}
	ctx.logger = nil

}
func (ctx *AlloterContext) GetImpl() interface{} {
	return ctx.Actx
}

//-ginRequest-------------------------------

type alloterRequest struct {
	actx   *alloter.Context
	vctx   *AlloterContext
	apath  *apath
	reqUrl *url.URL
	aquery *aquery
	abody  *abody
	closed bool
}

func (r *alloterRequest) ContentType() string {
	return r.actx.ContentType()
}

func (r *alloterRequest) GetMethod() string {
	return r.actx.Request.GetMethod()
}

func (r *alloterRequest) GetImpl() interface{} {
	return r.actx.Request
}

func (r *alloterRequest) RequestID() string {
	return r.vctx.Log().SessionID()
}

func (r *alloterRequest) GetClientIP() string {
	return r.actx.ClientIP()
}

func (r *alloterRequest) Header() vctx.Header {
	return xtypes.SMap(r.actx.Request.GetHeader())
}

func (r *alloterRequest) GetHeader(key string) string {
	return r.actx.GetHeader(key)
}

func (r *alloterRequest) SetHeader(key, val string) {
	r.actx.Header(key, val)
}

func (r *alloterRequest) Path() vctx.Path {
	if r.apath.closed {
		r.apath.closed = false
		r.apath.actx = r.actx
		r.apath.reqUrl = r.reqUrl
	}
	return r.apath
}

func (r *alloterRequest) Query() vctx.Query {
	if r.aquery.closed {
		r.aquery.closed = false
		r.aquery.actx = r.actx
		r.aquery.reqUrl = r.reqUrl
	}
	return r.aquery
}
func (r *alloterRequest) Body() vctx.Body {
	if r.abody.closed {
		r.abody.closed = false
		r.abody.actx = r.actx
		r.abody.vctx = r.vctx
	}
	return r.abody
}

func (q *alloterRequest) Close() {
	q.closed = true
	q.actx = nil
	q.vctx = nil
	q.reqUrl = nil
	q.apath.Close()
	q.aquery.Close()
	q.abody.Close()
}

//-path-------------------------

type apath struct {
	actx   *alloter.Context
	params xtypes.SMap
	reqUrl *url.URL
	closed bool
}

func (p *apath) GetURL() *url.URL {
	return p.reqUrl
}

func (p *apath) FullPath() string {
	return p.actx.FullPath()
}

func (p *apath) Params() xtypes.SMap {
	if p.params == nil {
		p.params = xtypes.SMap{}
		tps := p.actx.Params
		for i := range tps {
			p.params[tps[i].Key] = tps[i].Value
		}
	}
	return p.params
}
func (q *apath) Close() {
	q.closed = true
	q.actx = nil
	q.reqUrl = nil
	q.params = nil
}

//-gquery---------------------------------

type aquery struct {
	actx   *alloter.Context
	reqUrl *url.URL
	params xtypes.SMap
	closed bool
}

func (q *aquery) Get(name string) string {
	return q.reqUrl.Query().Get(name)
}
func (q *aquery) Values() xtypes.SMap {
	if q.params == nil {
		vals := q.reqUrl.Query()
		q.params = make(xtypes.SMap)
		for k := range vals {
			q.params[k] = vals.Get(k)
		}
	}
	return q.params
}
func (q *aquery) ScanTo(obj interface{}) error {
	return q.Values().ScanTo(obj)
}

func (q *aquery) String() string {
	return q.reqUrl.RawQuery
}

func (q *aquery) GetValues() url.Values {
	return q.reqUrl.Query()
}

func (q *aquery) Close() {
	q.actx = nil
	q.reqUrl = nil
	q.params = nil
	q.closed = true
}

// -gbody---------------------------------
type abody struct {
	actx      *alloter.Context
	vctx      *AlloterContext
	hasRead   bool
	reader    *bytes.Reader
	bodyBytes []byte
	closed    bool
}

func (q *abody) Scan(obj interface{}) error {
	return q.vctx.opts.RequestDecoder(q.vctx, obj)
}

func (q *abody) Read(p []byte) (n int, err error) {
	err = q.loadBody()
	if err != nil {
		return
	}
	return q.reader.Read(p)
}

func (q *abody) Len() int {
	err := q.loadBody()
	if err != nil {
		return 0
	}
	return len(q.bodyBytes)
}

func (q *abody) Bytes() []byte {
	err := q.loadBody()
	if err != nil {
		return nil
	}
	return q.bodyBytes
}

func (q *abody) loadBody() (err error) {
	if len(q.bodyBytes) == 0 && !q.hasRead {
		q.hasRead = true
		q.bodyBytes = q.actx.Request.Body()
		q.reader = bytes.NewReader(q.bodyBytes)
	}
	return nil
}

func (q *abody) Close() {
	q.bodyBytes = nil
	q.reader = nil
	q.actx = nil
	q.closed = true
	q.hasRead = false
}

// gresponse --------------------------------
type alloterResponse struct {
	vctx       *AlloterContext
	actx       *alloter.Context
	writebytes []byte
	hasWrited  bool
	closed     bool
	statusCode int
}

func (q *alloterResponse) Redirect(statusCode int, location string) {
	q.Status(statusCode)
	q.Header("Location", location)
}

func (q *alloterResponse) Status(statusCode int) {
	q.statusCode = statusCode
	q.actx.Writer.WriteHeader(statusCode)
}

func (q *alloterResponse) GetStatusCode() int {
	if q.statusCode == 0 {
		q.statusCode = http.StatusOK
	}
	return q.statusCode
}
func (q *alloterResponse) GetHeader(key string) string {
	return q.actx.Writer.Header().Get(key)
}

func (q *alloterResponse) Header(key, val string) {
	q.actx.Writer.Header().Set(key, val)
}

func (q *alloterResponse) ContextType(val string) {
	q.actx.Writer.Header().Set(constants.ContentTypeName, val)
}

func (q *alloterResponse) Write(obj interface{}) error {
	if q.hasWrited {
		panic(fmt.Errorf("%s:有多种方式写入响应", q.actx.FullPath()))
	}
	q.hasWrited = true
	if werr, ok := obj.(error); ok {
		q.vctx.opts.ErrorEncoder(q.vctx, werr)
		return nil
	}
	return q.vctx.opts.ResponseEncoder(q.vctx, obj)
}

func (q *alloterResponse) WriteBytes(bytes []byte) error {
	_, err := q.actx.Writer.Write(bytes)
	q.writebytes = bytes
	return err
}

func (q *alloterResponse) ContentType() string {
	return q.actx.Writer.Header().Get(constants.ContentTypeName)
}

func (q *alloterResponse) ResponseBytes() []byte {
	return q.writebytes
}
func (q *alloterResponse) Flush() error {
	return q.actx.Writer.Flush()
}
func (q *alloterResponse) Close() {
	q.vctx = nil
	q.actx = nil
	q.writebytes = nil
	q.hasWrited = false
	q.closed = true
	q.statusCode = http.StatusOK
}
