package ipc

import (
	"encoding/json"
	"testing"
)

func TestRequestBindEmptyDataKeepsZeroValue(t *testing.T) {
	req := &jsonRequest{cmd: "Reset"}
	args := echoArgs{Value: "unchanged"}

	if err := req.Bind(&args); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if args.Value != "unchanged" {
		t.Fatalf("Bind() value = %q, want unchanged value", args.Value)
	}
}

func TestRequestBindEmptyDataRejectsInvalidTarget(t *testing.T) {
	req := &jsonRequest{cmd: "Reset"}
	if err := req.Bind(nil); err == nil {
		t.Fatal("Bind() error = nil, want invalid target error")
	}
}

func TestResponseUnmarshalPreservesDataForTypedBind(t *testing.T) {
	const payload = `{"code":0,"data":{"value":"hello"}}`
	var response Response
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	var got echoArgs
	if err := response.Bind(&got); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if got.Value != "hello" {
		t.Fatalf("Bind() value = %q, want %q", got.Value, "hello")
	}
}

func TestResponseUnmarshalRejectsMalformedJSON(t *testing.T) {
	var response Response
	if err := json.Unmarshal([]byte(`{"code":`), &response); err == nil {
		t.Fatal("json.Unmarshal() error = nil, want malformed JSON error")
	}
}

func TestResponseBindUsesDataForLocallyCreatedResponse(t *testing.T) {
	response := OK(echoArgs{Value: "hello"})
	var got echoArgs

	if err := response.Bind(&got); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if got.Value != "hello" {
		t.Fatalf("Bind() value = %q, want %q", got.Value, "hello")
	}
}

func TestNilResponseBindReturnsError(t *testing.T) {
	var response *Response
	if err := response.Bind(&echoArgs{}); err == nil {
		t.Fatal("Bind() error = nil, want nil response error")
	}
}
