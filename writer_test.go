package jsontree

import (
	"bytes"
	"fmt"
	"testing"
)

func TestWriterWriteParent(t *testing.T) {
	// Calling WriteParent twice returns error
	{
		var buf bytes.Buffer
		w := NewWriter(&buf)
		if err := w.WriteParent("parent"); err != nil {
			t.Fatalf("First call to WriteParent() return error: %v", err)
		}
		l := buf.Len()
		want := fmt.Errorf("WriteParent() has already been called")
		if err := w.WriteParent("parent"); !errEqual(want, err) {
			t.Errorf("Duplicate call to WriteParent() return wrong error.\nWant %v\nGot  %v", want, err)
		} else if buf.Len() != l {
			t.Errorf("Duplicate call to WriteParent() wrote to the writer")
		}
	}
	// Calling WriteParent after WriteNode returns error
	{
		var buf bytes.Buffer
		w := NewWriter(&buf)
		node := &Node{Key: "k", Value: val("v")}
		if err := w.WriteNode(node); err != nil {
			t.Fatalf("WriteNode() return error: %v", err)
		}
		l := buf.Len()
		want := fmt.Errorf("WriteParent() must be called before any call to WriteNode()")
		if err := w.WriteParent("parent"); !errEqual(want, err) {
			t.Errorf("Calling WriteParent() after calling WriteNode() return wrong error.\nWant %v\nGot  %v", want, err)
		} else if buf.Len() != l {
			t.Errorf("Calling WriteParent() after calling WriteNode() ")
		}
	}
	// Calling WriteParent after Close() returns error
	{
		var buf bytes.Buffer
		w := NewWriter(&buf)
		if err := w.Close(); err != nil {
			t.Fatalf("Close() return error: %v", err)
		}
		l := buf.Len()
		want := fmt.Errorf("the writer is closed")
		if err := w.WriteParent("parent"); !errEqual(want, err) {
			t.Errorf("Calling WriteParent() after Close() return wrong error.\nWant %v\nGot  %v", want, err)
		} else if buf.Len() != l {
			t.Errorf("Calling WriteParent() after Close() wrote to the writer")
		}
	}
}

func TestWriterWriteNode(t *testing.T) {
	// Calling to closed writer return error
	{
		var buf bytes.Buffer
		w := NewWriter(&buf)
		if err := w.Close(); err != nil {
			t.Fatalf("Close() return error: %v", err)
		}
		l := buf.Len()
		node := &Node{Key: "k", Value: val("v")}
		want := fmt.Errorf("the writer is closed")
		if err := w.WriteNode(node); !errEqual(want, err) {
			t.Errorf("Calling WriteNode() after calling Close() return wrong error.\nWant %v\nGot  %v", want, err)
		} else if buf.Len() != l {
			t.Errorf("Calling WriteNode() after calling Close() wrote to the writer")
		}
	}
}

func TestWriter(t *testing.T) {
	tests := []struct {
		name   string
		parent string
		nodes  []*Node
		want   string
	}{
		{
			name:  "Empty slice => {}",
			nodes: nil,
			want:  "{}",
		},
		{
			name:   "Empty slice with parent",
			parent: "root",
			nodes:  nil,
			want:   `{"root":{}}`,
		},
		{
			name: "Normal nodes slice without parent",
			nodes: []*Node{
				&Node{Key: "a", Nodes: []*Node{
					&Node{Key: "b", Value: val("v1")},
					&Node{Key: "c", Nodes: []*Node{
						&Node{Key: "d", Value: val("v2")},
						&Node{Key: "e", Value: val("v3")},
					}},
				}},
				&Node{Key: "f", Nodes: []*Node{
					&Node{Key: "g", Value: val("v4")}}},
				&Node{Key: "h", Nodes: []*Node{
					&Node{Key: "i", Nodes: []*Node{
						&Node{Key: "j", Value: val("v5")},
					}},
				}},
			},
			want: `{"a":{"b":"v1","c":{"d":"v2","e":"v3"}},"f":{"g":"v4"},"h":{"i":{"j":"v5"}}}`,
		},
		{
			name:   "Normal nodes slice with parent",
			parent: "root",
			nodes: []*Node{
				&Node{Key: "a", Nodes: []*Node{
					&Node{Key: "b", Value: val("v1")},
					&Node{Key: "c", Nodes: []*Node{
						&Node{Key: "d", Value: val("v2")},
						&Node{Key: "e", Value: val("v3")},
					}},
				}},
				&Node{Key: "f", Nodes: []*Node{
					&Node{Key: "g", Value: val("v4")}}},
				&Node{Key: "h", Nodes: []*Node{
					&Node{Key: "i", Nodes: []*Node{
						&Node{Key: "j", Value: val("v5")},
					}},
				}},
			},
			want: `{"root":{"a":{"b":"v1","c":{"d":"v2","e":"v3"}},"f":{"g":"v4"},"h":{"i":{"j":"v5"}}}}`,
		},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		w := NewWriter(&buf)
		if test.parent != "" {
			if err := w.WriteParent(test.parent); err != nil {
				t.Fatalf("WriteParent() returned error: %v", err)
			}
		}
		for _, node := range test.nodes {
			if err := w.WriteNode(node); err != nil {
				t.Fatalf("WriteNode() returned error: %v", err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatalf("Close() returned error: %v", err)
		}
		if got := buf.String(); test.want != got {
			t.Errorf("Wrong data saved\nWant %v\nGot  %v", test.want, got)
		}
		// Writing a node using SerializeNode() and by calling Writer w with w.WriteParent(node.Key)
		// and then w.WriteNode() for each of node's Nodes should produce the same output
		if test.parent != "" {
			parent := &Node{Key: test.parent, Nodes: test.nodes}
			var buf2 bytes.Buffer
			if err := SerializeNode(parent, &buf2); err != nil {
				t.Fatalf("SerializeNode() failed: %v", err)
			} else if want, got := buf2.String(), buf.String(); want != got {
				t.Errorf("SerializeNode() and Writer do not produce the same result\nWant %v\nGot  %v", want, got)
			}
		}
	}
}
