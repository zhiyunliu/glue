package main

import (
	"github.com/zhiyunliu/velocity"
	"github.com/zhiyunliu/velocity/server/mqc"
)

func main() {
	app := velocity.NewApp()
	app.AddServer(mqc.New(""))
	app.Start()
}
