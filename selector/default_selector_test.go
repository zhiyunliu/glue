package selector

import (
	"context"
	"errors"
	"testing"
	"time"
)

type testNode struct {
	addr     string
	name     string
	version  string
	metadata map[string]string
	weight   *int64
}

func (n *testNode) Address() string        { return n.addr }
func (n *testNode) ServiceName() string    { return n.name }
func (n *testNode) InitialWeight() *int64  { return n.weight }
func (n *testNode) Version() string        { return n.version }
func (n *testNode) Metadata() map[string]string { return n.metadata }

type testWeightedNode struct {
	raw    Node
	weight float64
}

func (n *testWeightedNode) Address() string             { return n.raw.Address() }
func (n *testWeightedNode) ServiceName() string         { return n.raw.ServiceName() }
func (n *testWeightedNode) InitialWeight() *int64       { return n.raw.InitialWeight() }
func (n *testWeightedNode) Version() string             { return n.raw.Version() }
func (n *testWeightedNode) Metadata() map[string]string { return n.raw.Metadata() }
func (n *testWeightedNode) Raw() Node                   { return n.raw }
func (n *testWeightedNode) Weight() float64             { return n.weight }
func (n *testWeightedNode) Pick() DoneFunc              { return func(context.Context, DoneInfo) {} }
func (n *testWeightedNode) PickElapsed() time.Duration  { return 0 }

type testNodeBuilder struct{}

func (b *testNodeBuilder) Build(n Node) WeightedNode {
	return &testWeightedNode{raw: n, weight: 1}
}

type testBalancer struct {
	received []WeightedNode
	err      error
}

func (b *testBalancer) Pick(_ context.Context, nodes []WeightedNode) (WeightedNode, DoneFunc, error) {
	b.received = append([]WeightedNode(nil), nodes...)
	if b.err != nil {
		return nil, nil, b.err
	}
	selected := nodes[0]
	return selected, selected.Pick(), nil
}

func TestDefaultSelect_NoAppliedNodes(t *testing.T) {
	d := &Default{NodeBuilder: &testNodeBuilder{}, Balancer: &testBalancer{}}
	selected, done, err := d.Select(context.Background())
	if err != ErrNoAvailable {
		t.Fatalf("expected ErrNoAvailable, got %v", err)
	}
	if selected != nil || done != nil {
		t.Fatalf("expected nil selected/done, got %v/%v", selected, done)
	}
}

func TestDefaultApplyAndNodes(t *testing.T) {
	d := &Default{NodeBuilder: &testNodeBuilder{}, Balancer: &testBalancer{}}
	n1 := &testNode{addr: "node-1"}
	n2 := &testNode{addr: "node-2"}
	d.Apply([]Node{n1, n2})

	got := d.Nodes()
	if len(got) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(got))
	}
	if got[0].Address() != "node-1" || got[1].Address() != "node-2" {
		t.Fatalf("unexpected node addresses: %s, %s", got[0].Address(), got[1].Address())
	}
}

func TestDefaultSelect_WithDefaultAndOptionFilters(t *testing.T) {
	balancer := &testBalancer{}
	d := &Default{
		NodeBuilder: &testNodeBuilder{},
		Balancer:    balancer,
		Filters: []Filter{
			func(_ context.Context, nodes []Node) []Node {
				return nodes[:2]
			},
		},
	}
	n1 := &testNode{addr: "node-1"}
	n2 := &testNode{addr: "node-2"}
	n3 := &testNode{addr: "node-3"}
	d.Apply([]Node{n1, n2, n3})

	selected, done, err := d.Select(context.Background(), WithFilter(func(_ context.Context, nodes []Node) []Node {
		return nodes[1:]
	}))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if selected.Address() != "node-2" {
		t.Fatalf("expected node-2 selected, got %s", selected.Address())
	}
	if done == nil {
		t.Fatal("expected non-nil done func")
	}
	if len(balancer.received) != 1 || balancer.received[0].Address() != "node-2" {
		t.Fatalf("unexpected candidates passed to balancer: %+v", balancer.received)
	}
}

func TestDefaultSelect_EmptyAfterFilters(t *testing.T) {
	d := &Default{
		NodeBuilder: &testNodeBuilder{},
		Balancer:    &testBalancer{},
		Filters: []Filter{
			func(_ context.Context, _ []Node) []Node { return nil },
		},
	}
	d.Apply([]Node{&testNode{addr: "node-1"}})

	selected, done, err := d.Select(context.Background())
	if err != ErrNoAvailable {
		t.Fatalf("expected ErrNoAvailable, got %v", err)
	}
	if selected != nil || done != nil {
		t.Fatalf("expected nil selected/done, got %v/%v", selected, done)
	}
}

func TestDefaultSelect_BalancerError(t *testing.T) {
	wantErr := errors.New("pick failed")
	d := &Default{
		NodeBuilder: &testNodeBuilder{},
		Balancer:    &testBalancer{err: wantErr},
	}
	d.Apply([]Node{&testNode{addr: "node-1"}})

	selected, done, err := d.Select(context.Background())
	if err != wantErr {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if selected != nil || done != nil {
		t.Fatalf("expected nil selected/done, got %v/%v", selected, done)
	}
}
