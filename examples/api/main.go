package main

import (
	"github.com/zhiyunliu/velocity"
	"github.com/zhiyunliu/velocity/server/api"
)

func main() {
	app := velocity.NewApp()
	app.AddServer(api.New(""))
	app.Start()
}
