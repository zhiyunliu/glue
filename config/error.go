package config

import "errors"

var (
	ErrorUnchanged   = errors.New("data not changed")
	ErrWatcherClosed = errors.New("watcher closed")
)
