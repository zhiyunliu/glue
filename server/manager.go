/*
 * @Author: lwnmengjing
 * @Date: 2021/6/7 5:39 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/6/7 5:39 下午
 */

package server

import (
	"context"
)

type Manager interface {
	Name() string
	Add(...Runnable)
	Start(context.Context) error
}

type Runnable interface {

	// Start 启动
	Start(ctx context.Context) error

	// Attempt 是否允许启动
	Attempt() bool

	//String 格式化
	String() string
}
