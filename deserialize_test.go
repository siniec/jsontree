package jsontree

import (
	"bytes"
	"fmt"
	"testing"
)

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
			// Nodes are ordered non-alphabetically
			in: `{"root":{"b":"3","c":"1","a":"2"}}`,
			want: &Node{Key: "root", Nodes: []*Node{
				&Node{Key: "b", Value: "3"},
				&Node{Key: "c", Value: "1"},
				&Node{Key: "a", Value: "2"},
			}},
		},
		// Weird but valid input
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

func benchmarkNodeDeserialization(n int, b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := bytes.NewBuffer(benchmarks.deserialization.ins[n-1])
		if _, err := DeserializeNode(r); err != nil {
			b.Fatalf("Error: %v", err)
		}
	}
}

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
