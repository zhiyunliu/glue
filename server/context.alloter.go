package server

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xtypes"
	vctx "github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/contrib/alloter"
	"github.com/zhiyunliu/velocity/log"
)

type AlloterContext struct {
	Actx *alloter.Context
	opts *options

	areq   *alloterRequest
	aresp  *alloterResponse
	logger log.Logger
}

func (ctx *AlloterContext) reset(gctx *alloter.Context) {
	ctx.Actx = gctx
	ctx.areq = nil
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
	if ctx.areq == nil {
		reqUrl, _ := url.Parse(ctx.Actx.Request.GetService())
		ctx.areq = &alloterRequest{actx: ctx.Actx, vctx: ctx, reqUrl: reqUrl}
	}
	return ctx.areq
}
func (ctx *AlloterContext) Response() vctx.Response {
	if ctx.aresp == nil {
		ctx.aresp = &alloterResponse{actx: ctx.Actx, vctx: ctx}
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
	if ctx.areq.apath != nil {
		ctx.areq.apath.params = nil
		ctx.areq.apath.actx = nil
		ctx.areq.apath = nil
	}

	if ctx.areq.aquery != nil {
		ctx.areq.aquery.params = nil
		ctx.areq.aquery.actx = nil
		ctx.areq.aquery = nil
	}

	if ctx.areq.abody != nil {
		ctx.areq.abody.bodyBytes = nil
		ctx.areq.abody.actx = nil
		ctx.areq.abody = nil
	}

	if ctx.areq.apath != nil {
		ctx.areq.apath.params = nil
		ctx.areq.apath.actx = nil
		ctx.areq.apath = nil
	}

	if ctx.aresp != nil {
		ctx.aresp.actx = nil
		ctx.aresp = nil
	}

	ctx.areq.actx = nil
	ctx.Actx = nil

	ctx.logger.Close()
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
	if r.apath == nil {
		r.apath = &apath{actx: r.actx, reqUrl: r.reqUrl}
	}
	return r.apath
}

func (r *alloterRequest) Query() vctx.Query {
	if r.aquery == nil {
		r.aquery = &aquery{actx: r.actx, reqUrl: r.reqUrl}
	}
	return r.aquery
}
func (r *alloterRequest) Body() vctx.Body {
	if r.abody == nil {
		r.abody = &abody{actx: r.actx, vctx: r.vctx}
	}
	return r.abody
}

//-path-------------------------

type apath struct {
	actx   *alloter.Context
	params xtypes.SMap
	reqUrl *url.URL
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

//-gquery---------------------------------

type aquery struct {
	actx   *alloter.Context
	reqUrl *url.URL
	params xtypes.SMap
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

//-gbody---------------------------------
type abody struct {
	actx      *alloter.Context
	vctx      *AlloterContext
	hasRead   bool
	bodyBytes []byte
}

func (q *abody) Scan(obj interface{}) error {
	return q.vctx.opts.RequestDecoder(q.vctx, obj)
}

func (q *abody) Read(p []byte) (n int, err error) {
	err = q.loadBody()
	if err != nil {
		return
	}
	return bytes.NewReader(q.bodyBytes).Read(p)
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
	}
	return nil
}

//gresponse --------------------------------
type alloterResponse struct {
	vctx      *AlloterContext
	actx      *alloter.Context
	hasWrited bool
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
		panic(fmt.Errorf("%s：有多种方式写入响应", q.actx.FullPath()))
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
