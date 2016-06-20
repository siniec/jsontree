package jsontree

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestSerializeNode(t *testing.T) {
	node := &Node{
		Key: "data",
		Nodes: []*Node{
			&Node{
				Key: "1",
				Nodes: []*Node{
					&Node{Key: "a", Value: "v1"},
					&Node{Key: "b", Nodes: []*Node{&Node{Key: "i", Value: "v2"}}},
				},
			},
			&Node{
				Key: "2",
				Nodes: []*Node{
					&Node{Key: "a", Value: "v3"},
					&Node{Key: "b", Nodes: []*Node{&Node{Key: "i", Value: "v4"}}},
				},
			},
		},
	}
	var buf bytes.Buffer
	want := `{"data":{"1":{"a":"v1","b":{"i":"v2"}},"2":{"a":"v3","b":{"i":"v4"}}}}`

	// SerializeTo()
	if err := SerializeNode(node, &buf); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if got := buf.String(); want != got {
		t.Errorf("Wrong JSON written.\nWant %s\nGot  %s", want, got)
	}
}

// ========== Benchmarking ==========

var benchmarks = struct {
	serialization struct {
		nodes []*Node
	}
	deserialization struct {
		ins [][]byte
	}
}{}

func init() {
	const n = 5
	benchmarks.serialization.nodes = make([]*Node, n)
	benchmarks.deserialization.ins = make([][]byte, n)
	for i := 0; i < n; i++ {
		node := getTestNode(i, i)
		benchmarks.serialization.nodes[i] = node
		var buf bytes.Buffer
		if err := SerializeNode(node, &buf); err != nil {
			panic(err)
		} else {
			benchmarks.deserialization.ins[i] = buf.Bytes()
		}
	}
}

func getTestNode(width, depth int) *Node {
	var fn func(i, width, depth int) *Node
	fn = func(i, width, depth int) *Node {
		node := &Node{
			Key: fmt.Sprintf("%d_%d", depth, i),
		}
		if depth == 0 {
			node.Value = "NodeVal"
		} else {
			node.Nodes = make([]*Node, width)
			for i := 0; i < width; i++ {
				node.Nodes[i] = fn(i, width, depth-1)
			}
		}
		return node
	}
	return fn(0, width, depth)
}

func benchmarkNodeSerialization(n int, b *testing.B) {
	node := benchmarks.serialization.nodes[n-1]
	for i := 0; i < b.N; i++ {
		if err := SerializeNode(node, ioutil.Discard); err != nil {
			b.Fatalf("Error: %v", err)
		}
	}
}

func BenchmarkNodeSerialization1(b *testing.B) { benchmarkNodeSerialization(1, b) }
func BenchmarkNodeSerialization2(b *testing.B) { benchmarkNodeSerialization(2, b) }
func BenchmarkNodeSerialization3(b *testing.B) { benchmarkNodeSerialization(3, b) }
func BenchmarkNodeSerialization4(b *testing.B) { benchmarkNodeSerialization(4, b) }
func BenchmarkNodeSerialization5(b *testing.B) { benchmarkNodeSerialization(5, b) }
