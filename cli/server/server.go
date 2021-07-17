package server

//Server 服务器
type Server interface {
	Start() error
	Shutdown()
}
