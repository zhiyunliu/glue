package ratelimit

import "errors"

type Limiter interface {
	Allow() (DoneFunc, error)
}

// DoneFunc is done function.
type DoneFunc func(DoneInfo)

// DoneInfo is done info.
type DoneInfo struct {
	Err error
}

var (
	ErrNotAllow = errors.New("ratelimit not allow")
)
