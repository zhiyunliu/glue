package sre

import "time"

type options struct {
	Success float64       `json:"success"`
	Request int64         `json:"request"`
	Bucket  int           `json:"bucket"`
	Window  time.Duration `json:"window"`
}
