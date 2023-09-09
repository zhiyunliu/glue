package engine

import "net/http"

type HttpEngine interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	StaticFile(string, string)
	Static(string, string)
}
