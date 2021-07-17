package server

//ResponsiveServer 服务器
type ResponsiveServer interface {
	Start() error
	Shutdown()
}
