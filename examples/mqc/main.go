package main

import (
	"github.com/zhiyunliu/velocity"
	"github.com/zhiyunliu/velocity/server/mqc"
)

func main() {
	app := velocity.NewApp(velocity.Server(mqc.New("")))
	app.Start()
}
