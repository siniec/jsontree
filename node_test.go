package jsontree

import (
	"testing"
)

func TestNodeGet(t *testing.T) {
	node := &Node{
		Key: key("a"),
		Nodes: []*Node{
			{
				Key: key("1"),
				Nodes: []*Node{
					{Key: key("a"), Nodes: []*Node{{Key: key("i"), Value: val("v1")}}},
					{Key: key("b"), Value: val("v2")},
				},
			},
			{
				Key: key("2"),
				Nodes: []*Node{
					{Key: key("a"), Value: val("v3")},
				},
			},
		},
	}
	tests := []struct {
		path [][]byte
		want *Node
	}{
		{[][]byte{{'a'}}, nil},
		{[][]byte{{'1'}}, node.Nodes[0]},
		{[][]byte{{'1'}, {'a'}}, node.Nodes[0].Nodes[0]},
		{[][]byte{{'1'}, {'a'}, {'i'}}, node.Nodes[0].Nodes[0].Nodes[0]},
		{[][]byte{{'1'}, {'a'}, {'X'}}, nil},
		{[][]byte{{'1'}, {'b'}}, node.Nodes[0].Nodes[1]},
		{[][]byte{{'2'}}, node.Nodes[1]},
		{[][]byte{{'2'}, {'a'}}, node.Nodes[1].Nodes[0]},
	}
	for _, test := range tests {
		got := node.get(test.path...)
		if got != test.want {
			t.Fatalf("node.Get(%v) = %v, want %v", test.path, nodeString(got), nodeString(test.want))
		}
	}
}

func TestNodeGetOrAdd(t *testing.T) {
	node := &Node{Key: key("root")}

	if got := node.getOrAdd(); got != nil {
		t.Errorf("node.getOrAdd() != nil (%s)", nodeString(got))
	}

	keyA := []byte{'A'}
	nodeA := node.getOrAdd(keyA)
	if len(node.Nodes) != 1 || node.Nodes[0] != nodeA || !keyEqual(nodeA.Key, keyA) {
		t.Fatalf("node.getOrAdd(A)")
	}

	keyB := []byte{'B'}
	nodeB := node.getOrAdd(keyB)
	if len(node.Nodes) != 2 || node.Nodes[0] != nodeA || node.Nodes[1] != nodeB || !keyEqual(nodeB.Key, keyB) {
		t.Fatalf("node.getOrAdd(B)")
	}
	// getOrAdd for existing path returns original node
	if _nodeA := node.getOrAdd(keyA); nodeA != _nodeA {
		t.Fatalf("_nodeA != nodeA")
	}
	if _nodeB := node.getOrAdd(keyB); nodeB != _nodeB {
		t.Fatalf("_nodeB != nodeB")
	}

	keyA1 := []byte{'1'}
	nodeA1 := node.getOrAdd(keyA, keyA1)
	// root node wasn't affected
	if len(node.Nodes) != 2 || node.Nodes[0] != nodeA || node.Nodes[1] != nodeB {
		t.Fatalf("node.getOrAdd(B)")
	}
	// nodeA1 was added
	if len(nodeA.Nodes) != 1 || nodeA.Nodes[0] != nodeA1 || !keyEqual(nodeA1.Key, keyA1) {
		t.Fatalf("nodeA.getOrAdd(1)")
	}
	keyB1, keyB1i, keyB1ia := []byte{'1'}, []byte{'i'}, []byte{'a'}
	nodeB1ia := nodeB.getOrAdd(keyA1, keyB1i, keyB1ia)
	nodeB1 := nodeB.Nodes[0]
	if !keyEqual(nodeB1.Key, keyB1) {
		t.Errorf("nodeB1.Key = %s", nodeB1.Key)
	}
	nodeB1i := nodeB1.Nodes[0]
	if !keyEqual(nodeB1i.Key, keyB1i) {
		t.Errorf("nodeB1i.Key = %s", nodeB1i.Key)
	}
	if !keyEqual(nodeB1ia.Key, keyB1ia) {
		t.Errorf("nodeB1ia.Key = %s", nodeB1ia.Key)
	}

	if _nodeB1ia := nodeB1i.Nodes[0]; _nodeB1ia != nodeB1ia {
		t.Errorf("_nodeB1ia != nodeB1ia")
	}
}

// ========== Utility ==========

func key(key string) []byte {
	return []byte(key)
}
