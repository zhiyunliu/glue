/*
 * @Author: lwnmengjing
 * @Date: 2021/6/7 5:54 下午
 * @Last Modified by: lwnmengjing
 * @Last Modified time: 2021/6/7 5:54 下午
 */

package server

import (
	"os"
	"path"
	"time"
)

type Option func(*options)

type options struct {
	Name                    string
	gracefulShutdownTimeout time.Duration
}

func setDefaultOptions() options {
	return options{
		Name:                    path.Base(os.Args[0]),
		gracefulShutdownTimeout: 5 * time.Second,
	}
}

func WithName(name string) Option {
	return func(o *options) {
		o.Name = name
	}
}
