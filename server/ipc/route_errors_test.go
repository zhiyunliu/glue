package ipc

import (
	"context"
	"strings"
	"testing"

	gctx "github.com/zhiyunliu/glue/context"
)

func TestFrameworkRoutePanicAndEncodingFailure(t *testing.T) {
	s := New("errors", WithAddr(testEndpoint(t)))
	s.Handle("/panic", func(gctx.Context) interface{} { panic("route panic") })
	s.Handle("/encode", func(gctx.Context) interface{} { return make(chan int) })
	s.Handle("/ok", func(gctx.Context) interface{} { return "alive" })
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := Dial(s.Address())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for _, tc := range []struct {
		command string
		code    int
		message string
	}{{"/panic", 500, "route panic"}, {"/encode", 502, "unsupported type"}, {"/ok", 0, ""}} {
		response, err := c.Call(tc.command, nil)
		if err != nil {
			t.Fatal(err)
		}
		if response.Code != tc.code || !strings.Contains(response.Msg, tc.message) {
			t.Fatalf("%s = %#v", tc.command, response)
		}
	}
}
