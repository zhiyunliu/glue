package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/zhiyunliu/glue/server/ipc"
)

type valueResponse struct {
	Value int64 `json:"value"`
}

func main() {
	if err := run(); err != nil {
		log.Printf("IPC 客户端失败：%v", err)
	}
}

func run() (runErr error) {
	endpoint := "glue-ipc-example"
	if runtime.GOOS != "windows" {
		endpoint = filepath.Join(os.TempDir(), "glue-ipc-example.sock")
	}
	client, err := ipc.Dial(endpoint)
	if err != nil {
		return fmt.Errorf("连接 IPC 服务：%w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("关闭 IPC 客户端：%w", err))
		}
	}()

	response, err := client.Call("UpdateValue", map[string]int64{"value": 42})
	if err != nil {
		return fmt.Errorf("调用 UpdateValue：%w", err)
	}
	if !response.IsOK() {
		return fmt.Errorf("UpdateValue 返回错误：code=%d sub_code=%s msg=%s",
			response.Code, response.SubCode, response.Msg)
	}
	var updated valueResponse
	if err := response.Bind(&updated); err != nil {
		return fmt.Errorf("解析 UpdateValue 响应：%w", err)
	}
	fmt.Printf("[A] B 已更新为 %d\n", updated.Value)

	response, err = client.Call("Reset", nil)
	if err != nil {
		return fmt.Errorf("调用 Reset：%w", err)
	}
	if !response.IsOK() {
		return fmt.Errorf("Reset 返回错误：code=%d sub_code=%s msg=%s",
			response.Code, response.SubCode, response.Msg)
	}
	fmt.Println("[A] Reset 成功")

	response, err = client.Call("Get", nil)
	if err != nil {
		return fmt.Errorf("调用 Get：%w", err)
	}
	if !response.IsOK() {
		return fmt.Errorf("Get 返回错误：code=%d sub_code=%s msg=%s",
			response.Code, response.SubCode, response.Msg)
	}
	var current valueResponse
	if err := response.Bind(&current); err != nil {
		return fmt.Errorf("解析 Get 响应：%w", err)
	}
	fmt.Printf("[A] B 当前值为 %d\n", current.Value)

	response, err = client.Call("Nope", nil)
	if err != nil {
		return fmt.Errorf("调用未知命令：%w", err)
	}
	if response.Code != ipc.CodeUnknownCommand {
		return fmt.Errorf("未知命令 code=%d，期望 %d", response.Code, ipc.CodeUnknownCommand)
	}
	fmt.Printf("[A] 未知命令按预期失败：code=%d sub_code=%s\n", response.Code, response.SubCode)
	return nil
}
