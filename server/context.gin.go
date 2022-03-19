package server

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/url"

	"github.com/zhiyunliu/velocity/extlib/xtypes"
	"github.com/zhiyunliu/velocity/log"

	"github.com/gin-gonic/gin"
	vctx "github.com/zhiyunliu/velocity/context"
)

type GinContext struct {
	Gctx *gin.Context
	greq *ginRequest
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
	return nil
}
func (ctx *GinContext) Log() log.Logger {
	return log.New("")
}
func (ctx *GinContext) Close() {

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

func (r *ginRequest) Header(key string) string {
	return r.gctx.GetHeader(key)
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
func (q *gquery) XMap() xtypes.SMap {
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
	if len(q.bodyBytes) == 0 && !q.hasRead {
		q.hasRead = true
		q.bodyBytes, err = ioutil.ReadAll(q.gctx.Request.Body)
		q.gctx.Request.Body.Close()
	}
	return bytes.NewReader(q.bodyBytes).Read(p)
}
