package jsontree

import "testing"

func TestGetNode(t *testing.T) {
	node := &testNode{
		key: key("a"),
		nodes: []*testNode{
			{
				key: key("1"),
				nodes: []*testNode{
					{key: key("a"), nodes: []*testNode{{key: key("i"), value: val("v1")}}},
					{key: key("b"), value: val("v2")},
				},
			},
			{
				key: key("2"),
				nodes: []*testNode{
					{key: key("a"), value: val("v3")},
				},
			},
		},
	}
	tests := []struct {
		path [][]byte
		want *testNode
	}{
		{[][]byte{{'a'}}, nil},
		{[][]byte{{'1'}}, node.nodes[0]},
		{[][]byte{{'1'}, {'a'}}, node.nodes[0].nodes[0]},
		{[][]byte{{'1'}, {'a'}, {'i'}}, node.nodes[0].nodes[0].nodes[0]},
		{[][]byte{{'1'}, {'a'}, {'X'}}, nil},
		{[][]byte{{'1'}, {'b'}}, node.nodes[0].nodes[1]},
		{[][]byte{{'2'}}, node.nodes[1]},
		{[][]byte{{'2'}, {'a'}}, node.nodes[1].nodes[0]},
	}
	for _, test := range tests {
		got := getNode(node, test.path...)
		ok := true
		if test.want == nil {
			ok = got == nil
		} else {
			ok = nodeEqual(test.want, got)
		}
		if !ok {
			t.Errorf("node.Get(%v) = %v, want %v", test.path, nodeString(got), nodeString(test.want))
		}
	}
}

func TestNodeGetOrAdd(t *testing.T) {
	node := &testNode{key: key("root")}

	if got := getOrAddNode(node); got != nil {
		t.Errorf("getOrAddNode(node) != nil (%s)", nodeString(got))
	}

	keyA := []byte{'A'}
	nodeA := getOrAddNode(node, keyA)
	if len(node.Nodes()) != 1 || node.Nodes()[0] != nodeA || !keyEqual(nodeA.Key(), keyA) {
		t.Fatalf("getOrAddNode(node,A)")
	}

	keyB := []byte{'B'}
	nodeB := getOrAddNode(node, keyB)
	if len(node.Nodes()) != 2 || node.Nodes()[0] != nodeA || node.Nodes()[1] != nodeB || !keyEqual(nodeB.Key(), keyB) {
		t.Fatalf("getOrAddNode(node,B)")
	}
	// getOrAdd for existing path returns original node
	if _nodeA := getOrAddNode(node, keyA); nodeA != _nodeA {
		t.Fatalf("_nodeA != nodeA")
	}
	if _nodeB := getOrAddNode(node, keyB); nodeB != _nodeB {
		t.Fatalf("_nodeB != nodeB")
	}

	keyA1 := []byte{'1'}
	nodeA1 := getOrAddNode(node, keyA, keyA1)
	// root node wasn't affected
	if len(node.Nodes()) != 2 || node.Nodes()[0] != nodeA || node.Nodes()[1] != nodeB {
		t.Fatalf("getOrAddNode(node,B)")
	}
	// nodeA1 was added
	if len(nodeA.Nodes()) != 1 || nodeA.Nodes()[0] != nodeA1 || !keyEqual(nodeA1.Key(), keyA1) {
		t.Fatalf("nodeA.getOrAdd(1)")
	}
	keyB1, keyB1i, keyB1ia := []byte{'1'}, []byte{'i'}, []byte{'a'}
	nodeB1ia := getOrAddNode(nodeB, keyA1, keyB1i, keyB1ia)
	nodeB1 := nodeB.Nodes()[0]
	if !keyEqual(nodeB1.Key(), keyB1) {
		t.Errorf("nodeB1.Key = %s", nodeB1.Key())
	}
	nodeB1i := nodeB1.Nodes()[0]
	if !keyEqual(nodeB1i.Key(), keyB1i) {
		t.Errorf("nodeB1i.Key = %s", nodeB1i.Key())
	}
	if !keyEqual(nodeB1ia.Key(), keyB1ia) {
		t.Errorf("nodeB1ia.Key = %s", nodeB1ia.Key())
	}

	if _nodeB1ia := nodeB1i.Nodes()[0]; _nodeB1ia != nodeB1ia {
		t.Errorf("_nodeB1ia != nodeB1ia")
	}
}

// ========== Utility ==========

type testNode struct {
	key    []byte
	value  *testValue
	nodes  []*testNode
	_nodes [100]Node
}

func (n *testNode) Key() []byte {
	return n.key
}

func (n *testNode) SetKey(key []byte) {
	n.key = key
}

func (n *testNode) Value() Value {
	if n.value == nil {
		n.value = new(testValue)
	}
	return n.value
}

func (n *testNode) Nodes() []Node {
	for i, node := range n.nodes {
		n._nodes[i] = node
	}
	return n._nodes[:len(n.nodes)]
}

func (n *testNode) AddNode(key []byte) Node {
	node := &testNode{key: key}
	n.nodes = append(n.nodes, node)
	return node
}

func key(key string) []byte {
	return []byte(key)
}
