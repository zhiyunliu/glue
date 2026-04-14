package random

import (
	"context"
	"testing"
	"time"

	"github.com/zhiyunliu/glue/selector"
)

type testNode struct {
	addr string
}

func (n *testNode) Address() string             { return n.addr }
func (n *testNode) ServiceName() string         { return "" }
func (n *testNode) InitialWeight() *int64       { return nil }
func (n *testNode) Version() string             { return "" }
func (n *testNode) Metadata() map[string]string { return nil }

type testWeightedNode struct {
	selector.Node
	picked int
}

func (n *testWeightedNode) Raw() selector.Node          { return n.Node }
func (n *testWeightedNode) Weight() float64             { return 1 }
func (n *testWeightedNode) PickElapsed() time.Duration  { return 0 }
func (n *testWeightedNode) Pick() selector.DoneFunc {
	n.picked++
	return func(context.Context, selector.DoneInfo) {}
}

func TestBalancerPick_EmptyNodes(t *testing.T) {
	b := &Balancer{}
	selected, done, err := b.Pick(context.Background(), nil)
	if err != selector.ErrNoAvailable {
		t.Fatalf("expected ErrNoAvailable, got %v", err)
	}
	if selected != nil || done != nil {
		t.Fatalf("expected nil selected/done, got %v/%v", selected, done)
	}
}

func TestBalancerPick_SingleNode(t *testing.T) {
	n := &testWeightedNode{Node: &testNode{addr: "n1"}}
	b := &Balancer{}
	selected, done, err := b.Pick(context.Background(), []selector.WeightedNode{n})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if selected.Address() != "n1" {
		t.Fatalf("expected n1 selected, got %s", selected.Address())
	}
	if done == nil {
		t.Fatal("expected non-nil done")
	}
	if n.picked != 1 {
		t.Fatalf("expected Pick called once, got %d", n.picked)
	}
}
