package jsontree

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
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

func TestDeserializeNode(t *testing.T) {
	tests := []struct {
		in   string
		want *Node
		err  string
	}{
		{
			// Top level is leaf
			in:   `{"a":"b"}`,
			want: &Node{Key: "a", Value: "b"},
		},
		{
			// Sibling leaf nodes
			in: `{"root":{"a":"b","c":"d"}}`,
			want: &Node{Key: "root", Nodes: []*Node{
				&Node{Key: "a", Value: "b"},
				&Node{Key: "c", Value: "d"},
			}},
		},
		{
			// Sibling non-leaf nodes
			in: `{"root":{"a":{"a1":"v1"},"b":{"b1":"v2"}}}`,
			want: &Node{Key: "root", Nodes: []*Node{
				&Node{Key: "a", Nodes: []*Node{
					&Node{Key: "a1", Value: "v1"},
				}},
				&Node{Key: "b", Nodes: []*Node{
					&Node{Key: "b1", Value: "v2"},
				}},
			}},
		},
		{
			// Leaf nodes on different levels
			in: `{"root":{"a":"v1","b":{"b1":{"b11":"v3"},"b2":"v2"}}}`,
			want: &Node{Key: "root", Nodes: []*Node{
				&Node{Key: "a", Value: "v1"},
				&Node{Key: "b", Nodes: []*Node{
					&Node{Key: "b1", Nodes: []*Node{
						&Node{Key: "b11", Value: "v3"},
					}},
					&Node{Key: "b2", Value: "v2"},
				}},
			}},
		},
		{
			// Nodes are ordered as they are ordered in the input string
			in: `{"root":{"b":"3","c":"1","a":"2"}}`,
			want: &Node{Key: "root", Nodes: []*Node{
				&Node{Key: "b", Value: "3"},
				&Node{Key: "c", Value: "1"},
				&Node{Key: "a", Value: "2"},
			}},
		},
		// Handle weird but valid input
		{
			in: `{"ro\"ot":{"{a}":"\"hello\"","b}":"\\backslash\nnewline"}}`,
			want: &Node{Key: `ro\"ot`, Nodes: []*Node{
				&Node{Key: `{a}`, Value: `\"hello\"`},
				&Node{Key: `b}`, Value: `\\backslash\nnewline`},
			}},
		},
		// Handling invalid input
		// -- JSON syntax error
		{
			in:  `{"a":"b"},`,
			err: "expected end of input. Got ','",
		},
		{
			in:  `{"a":"b"`,
			err: "reader returned io.EOF before expected",
		},
		// -- Semantic error
		{
			in:  `{"a":"b","c":"d"}`,
			err: "invalid json. Expected 1 root node",
		},
	}
	for _, test := range tests {
		node, err := DeserializeNode(bytes.NewReader([]byte(test.in)))
		if test.err != "" {
			want := fmt.Errorf(test.err)
			if !errEqual(want, err) {
				t.Errorf("%s\nWrong error.\nWant %v\nGot  %v", test.in, want, err)
			}
			continue
		} else {
			if err != nil {
				t.Errorf("%s\nUnexpected error %v", test.in, err)
				continue
			}
		}
		if !nodeEqual(node, test.want) {
			t.Errorf("%s: Node was not as expected\nWant %v\nGot  %v", test.in, nodeString(test.want), nodeString(node))
		}
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

func benchmarkNodeDeserialization(n int, b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := bytes.NewBuffer(benchmarks.deserialization.ins[n-1])
		if _, err := DeserializeNode(r); err != nil {
			b.Fatalf("Error: %v", err)
		}
	}
}

func BenchmarkNodeSerialization1(b *testing.B) { benchmarkNodeSerialization(1, b) }
func BenchmarkNodeSerialization2(b *testing.B) { benchmarkNodeSerialization(2, b) }
func BenchmarkNodeSerialization3(b *testing.B) { benchmarkNodeSerialization(3, b) }
func BenchmarkNodeSerialization4(b *testing.B) { benchmarkNodeSerialization(4, b) }
func BenchmarkNodeSerialization5(b *testing.B) { benchmarkNodeSerialization(5, b) }

func BenchmarkNodeDeserialization1(b *testing.B) { benchmarkNodeDeserialization(1, b) }
func BenchmarkNodeDeserialization2(b *testing.B) { benchmarkNodeDeserialization(2, b) }
func BenchmarkNodeDeserialization3(b *testing.B) { benchmarkNodeDeserialization(3, b) }
func BenchmarkNodeDeserialization4(b *testing.B) { benchmarkNodeDeserialization(4, b) }
func BenchmarkNodeDeserialization5(b *testing.B) { benchmarkNodeDeserialization(5, b) }

// ========== Utility ==========

func nodeString(node *Node) string {
	if node == nil {
		return "<nil>"
	}
	var buf bytes.Buffer
	if err := SerializeNode(node, &buf); err != nil {
		return "<unserializable node>"
	} else {
		return buf.String()
	}
}

func nodesString(nodes []*Node) string {
	s := make([]string, len(nodes))
	for i, node := range nodes {
		s[i] = nodeString(node)
	}
	return "\t" + strings.Join(s, "\n\t")
}

func errEqual(want, got error) bool {
	return got != nil && want.Error() == got.Error()
}

func nodeEqual(want, got *Node) bool {
	if got == nil {
		return false
	}
	if got.Key != want.Key || got.Value != want.Value {
		return false
	}
	if gn, wn := got.Nodes, want.Nodes; len(gn) != len(wn) || (gn == nil && wn != nil) || (gn != nil && wn == nil) {
		return false
	}
	for i := range want.Nodes {
		if !nodeEqual(want.Nodes[i], got.Nodes[i]) {
			return false
		}
	}
	return true
}
