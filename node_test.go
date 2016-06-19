package jsontree

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestNodeSerialization(t *testing.T) {
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
	if err := node.SerializeTo(&buf); err != nil {
		t.Errorf("SerializeTo() returned error: %v", err)
	} else if got := buf.String(); want != got {
		t.Errorf("SerializeTo wrote wrong content.\nWant %s\nGot  %s", want, got)
	}

	// Serialize()
	if got, err := node.Serialize(); err != nil {
		t.Errorf("Serialize() returned error: %v", err)
	} else {
		if want != string(got) {
			t.Errorf("Serialize() returned wrong result.\nWant %s\nGot  %s", want, string(got))
		}
	}
}

func TestNodeDeserialization(t *testing.T) {
	tests := []struct {
		in   string
		want *Node
		err  string
	}{
		{
			in:   `{"a":"b"}`,
			want: &Node{Key: "a", Value: "b"},
		},
		{
			in:  `{"a":"b"},`,
			err: "invalid character ',' after top-level value", // original encoding/json package error
		},
		{
			in:  `{"a":"b","c":"d"}`,
			err: "Invalid json. Expected 1 root note. got 2",
		},
		{
			in: `{"root":{"a":"b"}}`,
			want: &Node{Key: "root", Nodes: []*Node{
				&Node{Key: "a", Value: "b"},
			}},
		},
		{
			in: `{"root":{"a":"b","c":"d"}}`,
			want: &Node{Key: "root", Nodes: []*Node{
				&Node{Key: "a", Value: "b"},
				&Node{Key: "c", Value: "d"},
			}},
		},
		{
			in:   `{"a}":{"\"b":"{v1"}}`,
			want: &Node{Key: "a}", Nodes: []*Node{&Node{Key: `"b`, Value: "{v1"}}},
		},
		{
			in:   `{"a":{"b":"v1","c":{"d":"v2","e":"v3"}}}`,
			want: &Node{Key: "a", Nodes: []*Node{&Node{Key: "b", Value: "v1"}, &Node{Key: "c", Nodes: []*Node{&Node{Key: "d", Value: "v2"}, &Node{Key: "e", Value: "v3"}}}}},
		},
	}
	for _, test := range tests {
		for _, stream := range []bool{true, false} {
			node := new(Node)
			var err error
			var name string
			if stream {
				name = fmt.Sprintf("DeserializeFrom(%s)", test.in)
				err = node.DeserializeFrom(bytes.NewReader([]byte(test.in)))
			} else {
				name = fmt.Sprintf("Deserialize(%s)", test.in)
				err = node.Deserialize([]byte(test.in))
			}
			if err != nil {
				if test.err == "" {
					t.Errorf(name+": Unexpected error %v", err)
				} else {
					if test.err != err.Error() {
						t.Errorf(name+": Wrong error.\nWant %v\nGot  %v", test.err, err)
					}
				}
			} else {
				if !nodeEqual(node, test.want) {
					t.Errorf(name+": Node was not as expected\nWant %v\nGot  %v", nodeString(test.want), nodeString(node))
				}
			}
		}
	}
}

// ========== Benchmarking ==========

var bench struct {
	result interface{} // used to prevent compiler optimization during benchmarks
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
	node := getTestNode(n, n)
	for i := 0; i < b.N; i++ {
		bench.result = node.SerializeTo(ioutil.Discard)
	}
}

func benchmarkNodeDeserialization(n int, b *testing.B) {
	node := getTestNode(n, n)
	bs, err := node.Serialize()
	if err != nil {
		b.Fatalf("Error serializing node: %v", err)
	}
	for i := 0; i < b.N; i++ {
		r := bytes.NewBuffer(bs)
		node := new(Node)
		bench.result = node.DeserializeFrom(r)
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
	if b, err := node.Serialize(); err != nil {
		return "<unserializable node>"
	} else {
		return string(b)
	}
}

func nodesString(nodes []*Node) string {
	s := make([]string, len(nodes))
	for i, node := range nodes {
		if node != nil {
			s[i] = nodeString(node)
		} else {
			s[i] = "<nil>"
		}
	}
	return "\t" + strings.Join(s, "\n\t")
}

func errEqual(want, got error) bool {
	return got != nil && want.Error() == got.Error()
}

func nodeEqual(want, got *Node) bool {
	if got.Key != want.Key || got.Value != want.Value {
		return false
	}
	if gn, wn := got.Nodes, want.Nodes; len(gn) != len(wn) || (gn == nil && wn != nil) || (gn != nil && wn == nil) {
		return false
	}
	// for i := range want.Nodes {
	// 	if !nodeEqual(want.Nodes[i], got.Nodes[i]) {
	// 		return false
	// 	}
	// }
	for _, wantChild := range want.Nodes {
		found := false
		for _, gotChild := range got.Nodes {
			if gotChild.Key == wantChild.Key {
				if !nodeEqual(wantChild, gotChild) {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
