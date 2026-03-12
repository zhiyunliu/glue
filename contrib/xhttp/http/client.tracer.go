package http

import (
	"net/http"
)

var _ http.RoundTripper = &TracerTransport{}

type TracerTransport struct {
	base http.RoundTripper
}

func NewTransport(base http.RoundTripper) *TracerTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	t := TracerTransport{
		base: base,
	}
	return &t
}

func (t *TracerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.base.RoundTrip(req)
	return res, err
}
