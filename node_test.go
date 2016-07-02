package jsontree

import (
	"testing"
)

func TestNodeGet(t *testing.T) {
	node := &Node{
		Key: "a",
		Nodes: []*Node{
			{
				Key: "1",
				Nodes: []*Node{
					{Key: "a", Nodes: []*Node{{Key: "i", Value: val("v1")}}},
					{Key: "b", Value: val("v2")},
				},
			},
			{
				Key: "2",
				Nodes: []*Node{
					{Key: "a", Value: val("v3")},
				},
			},
		},
	}
	tests := []struct {
		path []string
		want *Node
	}{
		{[]string{"a"}, nil},
		{[]string{"1"}, node.Nodes[0]},
		{[]string{"1", "a"}, node.Nodes[0].Nodes[0]},
		{[]string{"1", "a", "i"}, node.Nodes[0].Nodes[0].Nodes[0]},
		{[]string{"1", "a", "X"}, nil},
		{[]string{"1", "b"}, node.Nodes[0].Nodes[1]},
		{[]string{"2"}, node.Nodes[1]},
		{[]string{"2", "a"}, node.Nodes[1].Nodes[0]},
	}
	for _, test := range tests {
		got := node.get(test.path...)
		if got != test.want {
			t.Fatalf("node.Get(%v) = %v, want %v", test.path, nodeString(got), nodeString(test.want))
		}
	}
}

func TestNodeGetOrAdd(t *testing.T) {
	node := &Node{Key: "root"}

	if got := node.getOrAdd(); got != nil {
		t.Errorf("node.getOrAdd() != nil (%s)", nodeString(got))
	}

	nodeA := node.getOrAdd("A")
	if len(node.Nodes) != 1 || node.Nodes[0] != nodeA || nodeA.Key != "A" {
		t.Fatalf("node.getOrAdd(A)")
	}
	nodeB := node.getOrAdd("B")
	if len(node.Nodes) != 2 || node.Nodes[0] != nodeA || node.Nodes[1] != nodeB || nodeB.Key != "B" {
		t.Fatalf("node.getOrAdd(B)")
	}
	// getOrAdd for existing path returns original node
	if _nodeA := node.getOrAdd("A"); nodeA != _nodeA {
		t.Fatalf("_nodeA != nodeA")
	}
	if _nodeB := node.getOrAdd("B"); nodeB != _nodeB {
		t.Fatalf("_nodeB != nodeB")
	}

	nodeA1 := node.getOrAdd("A", "1")
	// root node wasn't affected
	if len(node.Nodes) != 2 || node.Nodes[0] != nodeA || node.Nodes[1] != nodeB {
		t.Fatalf("node.getOrAdd(B)")
	}
	// nodeA1 was added
	if len(nodeA.Nodes) != 1 || nodeA.Nodes[0] != nodeA1 || nodeA1.Key != "1" {
		t.Fatalf("nodeA.getOrAdd(1)")
	}
	nodeB1ia := nodeB.getOrAdd("1", "i", "a")
	nodeB1 := nodeB.Nodes[0]
	if nodeB1.Key != "1" {
		t.Errorf("nodeB1.Key = %s", nodeB1.Key)
	}
	nodeB1i := nodeB1.Nodes[0]
	if nodeB1i.Key != "i" {
		t.Errorf("nodeB1i.Key = %s", nodeB1i.Key)
	}
	if nodeB1ia.Key != "a" {
		t.Errorf("nodeB1ia.Key = %s", nodeB1ia.Key)
	}

	if _nodeB1ia := nodeB1i.Nodes[0]; _nodeB1ia != nodeB1ia {
		t.Errorf("_nodeB1ia != nodeB1ia")
	}
}
