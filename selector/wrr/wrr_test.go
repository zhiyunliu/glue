package wrr

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
	weight float64
	picked int
}

func (n *testWeightedNode) Raw() selector.Node         { return n.Node }
func (n *testWeightedNode) Weight() float64            { return n.weight }
func (n *testWeightedNode) PickElapsed() time.Duration { return 0 }
func (n *testWeightedNode) Pick() selector.DoneFunc {
	n.picked++
	return func(context.Context, selector.DoneInfo) {}
}

func TestBalancerPick_EmptyNodes(t *testing.T) {
	b := &Balancer{currentWeight: make(map[string]float64)}
	selected, done, err := b.Pick(context.Background(), nil)
	if err != selector.ErrNoAvailable {
		t.Fatalf("expected ErrNoAvailable, got %v", err)
	}
	if selected != nil || done != nil {
		t.Fatalf("expected nil selected/done, got %v/%v", selected, done)
	}
}

func TestBalancerPick_WeightedRoundRobinSequence(t *testing.T) {
	b := &Balancer{currentWeight: make(map[string]float64)}
	a := &testWeightedNode{Node: &testNode{addr: "a"}, weight: 5}
	c := &testWeightedNode{Node: &testNode{addr: "c"}, weight: 1}
	nodes := []selector.WeightedNode{a, c}

	want := []string{"a", "a", "a", "c"}
	for i, w := range want {
		selected, done, err := b.Pick(context.Background(), nodes)
		if err != nil {
			t.Fatalf("pick %d: expected nil err, got %v", i, err)
		}
		if selected.Address() != w {
			t.Fatalf("pick %d: expected %s, got %s", i, w, selected.Address())
		}
		if done == nil {
			t.Fatalf("pick %d: expected non-nil done", i)
		}
	}

	if a.picked != 3 || c.picked != 1 {
		t.Fatalf("unexpected pick counts, a=%d c=%d", a.picked, c.picked)
	}
}
