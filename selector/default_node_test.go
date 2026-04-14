package selector

import (
	"testing"

	"github.com/zhiyunliu/glue/registry"
)

func TestNewNode_WithNilInstance(t *testing.T) {
	n := NewNode(registry.ServerItem{EndpointURL: "http://127.0.0.1:8080/path"}, nil)
	dn, ok := n.(*DefaultNode)
	if !ok {
		t.Fatalf("expected *DefaultNode, got %T", n)
	}
	if dn.Address() != "http://127.0.0.1:8080/path" {
		t.Fatalf("unexpected address: %s", dn.Address())
	}
	if dn.ServiceName() != "" || dn.Version() != "" {
		t.Fatalf("expected empty name/version, got %s/%s", dn.ServiceName(), dn.Version())
	}
	if dn.InitialWeight() != nil {
		t.Fatalf("expected nil weight, got %v", *dn.InitialWeight())
	}
	if dn.Scheme() != "http" || dn.Host() != "127.0.0.1" || dn.Port() != "8080" {
		t.Fatalf("unexpected parsed addr info: %s %s %s", dn.Scheme(), dn.Host(), dn.Port())
	}
}

func TestNewNode_MetadataWeightOverride(t *testing.T) {
	n := NewNode(
		registry.ServerItem{EndpointURL: "grpc://service.local:9000"},
		&registry.ServiceInstance{
			Name:    "svc",
			Version: "v1",
			Weight:  10,
			Metadata: map[string]string{
				"weight": "25",
				"zone":   "sh",
			},
		},
	)
	dn := n.(*DefaultNode)

	if dn.ServiceName() != "svc" || dn.Version() != "v1" {
		t.Fatalf("unexpected service info: %s/%s", dn.ServiceName(), dn.Version())
	}
	if dn.InitialWeight() == nil || *dn.InitialWeight() != 25 {
		t.Fatalf("expected overridden weight 25, got %+v", dn.InitialWeight())
	}
	if dn.Metadata()["zone"] != "sh" {
		t.Fatalf("expected metadata zone=sh, got %+v", dn.Metadata())
	}
	if dn.Scheme() != "grpc" || dn.Host() != "service.local" || dn.Port() != "9000" {
		t.Fatalf("unexpected parsed addr info: %s %s %s", dn.Scheme(), dn.Host(), dn.Port())
	}
}

func TestNewNode_InvalidMetadataWeightFallback(t *testing.T) {
	n := NewNode(
		registry.ServerItem{EndpointURL: "http://localhost:18080"},
		&registry.ServiceInstance{
			Weight: 12,
			Metadata: map[string]string{
				"weight": "not-number",
			},
		},
	)
	dn := n.(*DefaultNode)
	if dn.InitialWeight() == nil || *dn.InitialWeight() != 12 {
		t.Fatalf("expected fallback weight 12, got %+v", dn.InitialWeight())
	}
}
