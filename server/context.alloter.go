package server

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	vctx "github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/contrib/alloter"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xtypes"
)

type AlloterContext struct {
	Actx *alloter.Context
	opts *options

	areq   *alloterRequest
	aresp  *alloterResponse
	logger log.Logger
}

func newAlloterContext(opts *options) *AlloterContext {
	return &AlloterContext{
		opts: opts,
		areq: &alloterRequest{
			hasClose: true,
			apath:    &apath{hasClose: true},
			aquery:   &aquery{hasClose: true},
			abody:    &abody{hasClose: true},
		},
		aresp: &alloterResponse{
			hasClose: true,
		},
	}
}

func (ctx *AlloterContext) reset(gctx *alloter.Context) {
	ctx.Actx = gctx
}

func (ctx *AlloterContext) ServerType() string {
	return ctx.opts.SrvType
}
func (ctx *AlloterContext) Context() context.Context {
	return ctx.Actx.Request.Context()
}

func (ctx *AlloterContext) ResetContext(nctx context.Context) {
	req := ctx.Actx.Request.WithContext(nctx)
	ctx.Actx.Request = req
}

func (ctx *AlloterContext) Header(key string) string {
	return ctx.Actx.GetHeader(key)
}

func (ctx *AlloterContext) Request() vctx.Request {
	if ctx.areq.hasClose {
		ctx.areq.hasClose = false
		reqUrl, _ := url.Parse(ctx.Actx.Request.GetService())
		ctx.areq.actx = ctx.Actx
		ctx.areq.vctx = ctx
		ctx.areq.reqUrl = reqUrl
	}
	return ctx.areq
}
func (ctx *AlloterContext) Response() vctx.Response {
	if ctx.aresp.hasClose {
		ctx.aresp.hasClose = false
		ctx.aresp.actx = ctx.Actx
		ctx.aresp.vctx = ctx
	}
	return ctx.aresp
}
func (ctx *AlloterContext) Log() log.Logger {
	if ctx.logger == nil {
		logger, ok := log.FromContext(ctx.Context())
		if !ok {
			xreqId := ctx.Actx.GetHeader(XRequestId)
			if xreqId == "" {
				xreqId = session.Create()
				ctx.Actx.Header(XRequestId, xreqId)
			}
			logger = log.New(log.WithName("alloter"), log.WithSid(xreqId))
			ctx.ResetContext(log.WithContext(ctx.Context(), logger))
		}
		ctx.logger = logger
	}
	return ctx.logger
}
func (ctx *AlloterContext) Close() {
	ctx.areq.Close()
	ctx.aresp.Close()
	ctx.Actx = nil

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
	actx     *alloter.Context
	vctx     *AlloterContext
	apath    *apath
	reqUrl   *url.URL
	aquery   *aquery
	abody    *abody
	hasClose bool
}

func (r *alloterRequest) GetMethod() string {
	return r.actx.Request.GetMethod()
}

func (r *alloterRequest) GetClientIP() string {
	return r.actx.ClientIP()
}

func (r *alloterRequest) GetHeader(key string) string {
	return r.actx.GetHeader(key)
}

func (r *alloterRequest) SetHeader(key, val string) {
	r.actx.Header(key, val)
}

func (r *alloterRequest) Path() vctx.Path {
	if r.apath.hasClose {
		r.apath.actx = r.actx
		r.apath.reqUrl = r.reqUrl
	}
	return r.apath
}

func (r *alloterRequest) Query() vctx.Query {
	if r.aquery.hasClose {
		r.aquery.actx = r.actx
		r.aquery.reqUrl = r.reqUrl
	}
	return r.aquery
}
func (r *alloterRequest) Body() vctx.Body {
	if r.abody.hasClose {
		r.abody.actx = r.actx
		r.abody.vctx = r.vctx
	}
	return r.abody
}

func (q *alloterRequest) Close() {
	q.hasClose = true
	q.actx = nil
	q.vctx = nil
	q.reqUrl = nil
	q.apath.Close()
	q.aquery.Close()
	q.abody.Close()
}

//-path-------------------------

type apath struct {
	actx     *alloter.Context
	params   xtypes.SMap
	reqUrl   *url.URL
	hasClose bool
}

func (p *apath) GetURL() *url.URL {
	return p.reqUrl
}

func (p *apath) FullPath() string {
	return p.actx.Request.GetService()
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
	q.hasClose = true
	q.actx = nil
	q.reqUrl = nil
	q.params = nil
}

//-gquery---------------------------------

type aquery struct {
	actx     *alloter.Context
	reqUrl   *url.URL
	params   xtypes.SMap
	hasClose bool
}

func (q *aquery) Get(name string) string {
	return q.reqUrl.Query().Get(name)
}
func (q *aquery) SMap() xtypes.SMap {
	if q.params == nil {
		vals := q.reqUrl.Query()
		q.params = make(xtypes.SMap)
		for k := range vals {
			q.params[k] = vals.Get(k)
		}
	}
	return q.params
}
func (q *aquery) Scan(obj interface{}) error {
	return q.SMap().Scan(obj)
}

func (q *aquery) String() string {
	return q.reqUrl.RawQuery
}

func (q *aquery) Close() {
	q.actx = nil
	q.reqUrl = nil
	q.params = nil
	q.hasClose = true
}

//-gbody---------------------------------
type abody struct {
	actx      *alloter.Context
	vctx      *AlloterContext
	hasRead   bool
	reader    *bytes.Reader
	bodyBytes []byte
	hasClose  bool
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
		if err != nil {
			return err
		}
		q.reader = bytes.NewReader(q.bodyBytes)
	}
	return nil
}

func (q *abody) Close() {
	q.bodyBytes = nil
	q.reader = nil
	q.actx = nil
	q.hasClose = true
	q.hasRead = false
}

//gresponse --------------------------------
type alloterResponse struct {
	vctx      *AlloterContext
	actx      *alloter.Context
	hasWrited bool
	hasClose  bool
}

func (q *alloterResponse) Status(statusCode int) {
	q.actx.Writer.WriteHeader(statusCode)
}

func (q *alloterResponse) Header(key, val string) {
	q.actx.Writer.Header().Set(key, val)
}

func (q *alloterResponse) ContextType(val string) {
	q.actx.Writer.Header().Set("content-type", val)
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
	return err
}
func (q *alloterResponse) Close() {
	q.vctx = nil
	q.actx = nil
	q.hasWrited = false
	q.hasClose = true
}
