package engine

import "net/http"

type HttpEngine interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
	StaticFS(relativePath string, fs http.FileSystem)
	StaticFile(string, string)
	Static(string, string)
}
