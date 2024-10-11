package api

import "net/http"

type Static struct {
	RouterPath string
	FilePath   string
	DirPath    string
	FileSystem http.FileSystem
}
