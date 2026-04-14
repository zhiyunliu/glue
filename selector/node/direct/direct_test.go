package direct

import (
	"testing"

	"github.com/zhiyunliu/glue/selector"
)

type testNode struct {
	addr   string
	weight *int64
}

func (n *testNode) Address() string             { return n.addr }
func (n *testNode) ServiceName() string         { return "" }
func (n *testNode) InitialWeight() *int64       { return n.weight }
func (n *testNode) Version() string             { return "" }
func (n *testNode) Metadata() map[string]string { return nil }

func TestNodeWeight_DefaultAndConfigured(t *testing.T) {
	wn := (&Builder{}).Build(&testNode{addr: "n1"})
	if wn.Weight() != 100 {
		t.Fatalf("expected default weight 100, got %v", wn.Weight())
	}

	w := int64(320)
	wn2 := (&Builder{}).Build(&testNode{addr: "n2", weight: &w})
	if wn2.Weight() != 320 {
		t.Fatalf("expected configured weight 320, got %v", wn2.Weight())
	}
}

func TestNodePickAndRaw(t *testing.T) {
	raw := &testNode{addr: "n1"}
	wn := (&Builder{}).Build(raw).(*Node)

	if wn.lastPick != 0 {
		t.Fatalf("expected initial lastPick=0, got %d", wn.lastPick)
	}

	done := wn.Pick()
	if done == nil {
		t.Fatal("expected non-nil done func")
	}
	if wn.lastPick == 0 {
		t.Fatal("expected lastPick updated after Pick")
	}

	if wn.Raw() != selector.Node(raw) {
		t.Fatalf("expected raw node unchanged, got %v", wn.Raw())
	}
}
