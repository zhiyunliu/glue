package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

type echoArgs struct {
	Value string `json:"value"`
}

func startTestServer(t *testing.T) (*Server, *Client) {
	t.Helper()
	server, err := NewServer(testEndpoint(t))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	serveResult := make(chan error, 1)
	go func() {
		serveResult <- server.Serve()
	}()

	client, err := Dial(server.Address())
	if err != nil {
		_ = server.Close()
		t.Fatalf("Dial() error = %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil && !errors.Is(err, ErrClosed) {
			t.Errorf("client Close() error = %v", err)
		}
		if err := server.Close(); err != nil {
			t.Errorf("server Close() error = %v", err)
		}
		select {
		case err := <-serveResult:
			if err != nil {
				t.Errorf("Serve() error after Close() = %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("Serve() did not return after Close()")
		}
	})
	return server, client
}

func TestClientCallBindsRequestAndResponse(t *testing.T) {
	server, client := startTestServer(t)
	server.RegisterFunc("Echo", func(req Request) *Response {
		var args echoArgs
		if err := req.Bind(&args); err != nil {
			return WrapError(CodeBadRequest, "bad_args", err)
		}
		return OK(args)
	})

	response, err := client.Call("Echo", echoArgs{Value: "hello"})
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if !response.IsOK() {
		t.Fatalf("Call() response = %#v, want success", response)
	}
	var got echoArgs
	if err := response.Bind(&got); err != nil {
		t.Fatalf("Response.Bind() error = %v", err)
	}
	if got.Value != "hello" {
		t.Fatalf("Response.Bind() value = %q, want %q", got.Value, "hello")
	}
}

func TestClientCallReturnsUnknownCommandResponse(t *testing.T) {
	_, client := startTestServer(t)

	response, err := client.Call("Missing", nil)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if response.Code != CodeUnknownCommand {
		t.Fatalf("Call() code = %d, want %d", response.Code, CodeUnknownCommand)
	}
}

func TestServerConvertsHandlerPanicToResponse(t *testing.T) {
	server, client := startTestServer(t)
	server.RegisterFunc("Panic", func(Request) *Response {
		panic("handler failed")
	})

	response, err := client.Call("Panic", nil)
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if response.Code != CodeHandlerPanic {
		t.Fatalf("Call() code = %d, want %d", response.Code, CodeHandlerPanic)
	}
}

func TestServerReturnsBadRequestForMalformedJSON(t *testing.T) {
	server, err := NewServer(testEndpoint(t))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	serveResult := make(chan error, 1)
	go func() {
		serveResult <- server.Serve()
	}()
	defer func() {
		if err := server.Close(); err != nil {
			t.Errorf("server Close() error = %v", err)
		}
		select {
		case err := <-serveResult:
			if err != nil {
				t.Errorf("Serve() error after Close() = %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("Serve() did not return after Close()")
		}
	}()

	conn, err := dialTransport(server.Address())
	if err != nil {
		t.Fatalf("dialTransport() error = %v", err)
	}
	defer conn.Close()
	if err := writeFrame(conn, []byte("{")); err != nil {
		t.Fatalf("writeFrame() error = %v", err)
	}
	payload, err := readFrame(conn)
	if err != nil {
		t.Fatalf("readFrame() error = %v", err)
	}
	var response Response
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.Code != CodeBadRequest {
		t.Fatalf("response code = %d, want %d", response.Code, CodeBadRequest)
	}
}

func TestServerCloseClosesActiveConnections(t *testing.T) {
	server, err := NewServer(testEndpoint(t))
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	serveResult := make(chan error, 1)
	go func() {
		serveResult <- server.Serve()
	}()

	conn, err := dialTransport(server.Address())
	if err != nil {
		_ = server.Close()
		t.Fatalf("dialTransport() error = %v", err)
	}
	defer conn.Close()

	if err := server.Close(); err != nil {
		t.Fatalf("server Close() error = %v", err)
	}
	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil &&
		!errors.Is(err, ErrDeadlineUnsupported) {
		t.Fatalf("SetReadDeadline() error = %v", err)
	}
	if _, err := readFrame(conn); err == nil {
		t.Fatal("readFrame() error = nil, want server connection close")
	}
	select {
	case err := <-serveResult:
		if err != nil {
			t.Fatalf("Serve() error after Close() = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve() did not return after Close()")
	}
}

func TestClientCallSerializesConcurrentRequests(t *testing.T) {
	server, client := startTestServer(t)
	server.RegisterFunc("Echo", func(req Request) *Response {
		var args echoArgs
		if err := req.Bind(&args); err != nil {
			return WrapError(CodeBadRequest, "bad_args", err)
		}
		return OK(args)
	})

	const calls = 8
	errs := make(chan error, calls)
	for i := 0; i < calls; i++ {
		go func(i int) {
			response, err := client.Call("Echo", echoArgs{Value: fmt.Sprintf("%d", i)})
			if err != nil {
				errs <- err
				return
			}
			var got echoArgs
			if err := response.Bind(&got); err != nil {
				errs <- err
				return
			}
			if got.Value != fmt.Sprintf("%d", i) {
				errs <- fmt.Errorf("response value = %q, want %q", got.Value, fmt.Sprintf("%d", i))
				return
			}
			errs <- nil
		}(i)
	}
	for range calls {
		if err := <-errs; err != nil {
			t.Errorf("concurrent Call() error = %v", err)
		}
	}
}

func TestClientRejectsCallAfterClose(t *testing.T) {
	_, client := startTestServer(t)
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if _, err := client.Call("Nope", nil); !errors.Is(err, ErrClosed) {
		t.Fatalf("Call() error = %v, want %v", err, ErrClosed)
	}
}
