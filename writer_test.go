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

func TestWriterClose(t *testing.T) {
	tests := []struct {
		name   string
		node   bool
		parent bool
		closed bool
		want   string
	}{
		{
			name:   "Closing closed writer does nothing",
			closed: true,
			want:   "",
		},
		{
			name: "No nodes or parents have been written",
			want: "{}",
		},
		{
			name: "Only nodes have been written",
			node: true,
			want: "}",
		},
		{
			name:   "Only parent has been written",
			parent: true,
			want:   "{}}",
		},
		{
			name:   "Parent and nodes have been written",
			node:   true,
			parent: true,
			want:   "}}",
		},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		w := NewWriter(&buf)
		w.closed, w.hasWrittenNode, w.hasWrittenParent = test.closed, test.node, test.parent
		if err := w.Close(); err != nil {
			t.Errorf("%s: Close() returned error: %v", test.name, err)
		}
		if buf.String() != test.want {
			t.Errorf("%s: Close() on closed writer wrote wrong contents\nWant %v\nGot  %v", test.name, test.want, buf.String())
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
				{Key: "a", Nodes: []*Node{
					{Key: "b", Value: val("v1")},
					{Key: "c", Nodes: []*Node{
						{Key: "d", Value: val("v2")},
						{Key: "e", Value: val("v3")},
					}},
				}},
				{Key: "f", Nodes: []*Node{
					{Key: "g", Value: val("v4")}}},
				{Key: "h", Nodes: []*Node{
					{Key: "i", Nodes: []*Node{
						{Key: "j", Value: val("v5")},
					}},
				}},
			},
			want: `{"a":{"b":"v1","c":{"d":"v2","e":"v3"}},"f":{"g":"v4"},"h":{"i":{"j":"v5"}}}`,
		},
		{
			name:   "Normal nodes slice with parent",
			parent: "root",
			nodes: []*Node{
				{Key: "a", Nodes: []*Node{
					{Key: "b", Value: val("v1")},
					{Key: "c", Nodes: []*Node{
						{Key: "d", Value: val("v2")},
						{Key: "e", Value: val("v3")},
					}},
				}},
				{Key: "f", Nodes: []*Node{
					{Key: "g", Value: val("v4")}}},
				{Key: "h", Nodes: []*Node{
					{Key: "i", Nodes: []*Node{
						{Key: "j", Value: val("v5")},
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
				t.Fatalf("%s: WriteParent() returned error: %v", test.name, err)
			}
		}
		for _, node := range test.nodes {
			if err := w.WriteNode(node); err != nil {
				t.Fatalf("%s: WriteNode() returned error: %v", test.name, err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatalf("%s: Close() returned error: %v", test.name, err)
		}
		if got := buf.String(); test.want != got {
			t.Errorf("%s: Wrong data saved\nWant %v\nGot  %v", test.name, test.want, got)
		}
		// Writing a node using SerializeNode() and by calling Writer w with w.WriteParent(node.Key)
		// and then w.WriteNode() for each of node's Nodes should produce the same output
		if test.parent != "" {
			parent := &Node{Key: test.parent, Nodes: test.nodes}
			var buf2 bytes.Buffer
			if err := SerializeNode(parent, &buf2); err != nil {
				t.Fatalf("%s: SerializeNode() failed: %v", test.name, err)
			} else if want, got := buf2.String(), buf.String(); want != got {
				t.Errorf("%s: SerializeNode() and Writer do not produce the same result\nWant %v\nGot  %v", test.name, want, got)
			}
		}

		// Test writer error
		for i := 0; i <= len(test.want); i++ {
			wantErr := fmt.Errorf("Test err")
			ew := &errWriter{
				errIndex: i,
				err:      wantErr,
			}
			w := NewWriter(ew)
			// Write until we encounter an error
			var gotErr error
			if test.parent != "" {
				gotErr = w.WriteParent(test.parent)
			}
			if gotErr == nil {
				for _, node := range test.nodes {
					if gotErr = w.WriteNode(node); gotErr != nil {
						break
					}
				}
			}
			if gotErr == nil {
				gotErr = w.Close()
			}

			if !errEqual(wantErr, gotErr) {
				t.Errorf("%s: (errWriter(%d)) Wrong error returned\nWant %v\nGot  %v", test.name, i, wantErr, gotErr)
			}
		}

		// Test node serialization error
		if len(test.nodes) > 0 {
			if leaf := findLeafNode(test.nodes); leaf != nil {
				// Set one of the nodes' values to return an error when serializing
				wantErr := fmt.Errorf("Serialize test err")
				leaf.Value = valErr("", wantErr, nil)
				var buf bytes.Buffer
				w := NewWriter(&buf)
				for _, node := range test.nodes {
					if err := w.WriteNode(node); err != nil {
						if !errEqual(wantErr, err) {
							t.Errorf("%s: (serialization error) Wrong error returned\nWant %v\nGot  %v", test.name, wantErr, err)
						}
						break
					}

				}
			}
		}
	}
}

func findLeafNode(nodes []*Node) *Node {
	for _, node := range nodes {
		if node.Value != nil {
			return node
		}
	}
	for _, node := range nodes {
		if leaf := findLeafNode(node.Nodes); leaf != nil {
			return leaf
		}
	}
	return nil
}
