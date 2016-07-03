package jsontree

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func TestNewWriter(t *testing.T) {
	// If w implements io.Writer and io.ByteWriter, use it directly
	{
		var buf bytes.Buffer
		w := NewWriter(&buf)
		if w.w != &buf {
			t.Errorf("NewWriter(bytes.Buffer) set wrong writer")
		}
	}
	// If w does not implement io.Writer or io.ByteWriter, wrap it in a bufio.Writer
	{
		mw := new(memWriter)
		w := NewWriter(mw)
		if bw, ok := w.w.(*bufio.Writer); !ok {
			t.Errorf("NewWriter() did not set internal writer to bufio.Writer")
		} else if err := bw.WriteByte('?'); err != nil {
			t.Errorf("bw.WriteByte() error: %v", err)
		} else if err = bw.Flush(); err != nil {
			t.Fatalf("bw.Flush() error: %v", err)
		} else if string(mw.bs) != "?" {
			t.Errorf("NewWriter() with bufio.Writer did not write to correct underlying writer")
		}
	}
}

func TestWriterWriteParent(t *testing.T) {
	parentKey := key("parent")
	// Calling WriteParent twice returns error
	{
		var buf bytes.Buffer
		w := NewWriter(&buf)
		if err := w.WriteParent(parentKey); err != nil {
			t.Fatalf("First call to WriteParent() return error: %v", err)
		}
		l := buf.Len()
		want := fmt.Errorf("WriteParent() has already been called")
		if err := w.WriteParent(parentKey); !errEqual(want, err) {
			t.Errorf("Duplicate call to WriteParent() return wrong error.\nWant %v\nGot  %v", want, err)
		} else if buf.Len() != l {
			t.Errorf("Duplicate call to WriteParent() wrote to the writer")
		}
	}
	// Calling WriteParent after WriteNode returns error
	{
		var buf bytes.Buffer
		w := NewWriter(&buf)
		node := &testNode{key: key("k"), value: val("v")}
		if err := w.WriteNode(node); err != nil {
			t.Fatalf("WriteNode() return error: %v", err)
		}
		l := buf.Len()
		want := fmt.Errorf("WriteParent() must be called before any call to WriteNode()")
		if err := w.WriteParent(parentKey); !errEqual(want, err) {
			t.Errorf("Calling WriteParent() after calling WriteNode() return wrong error.\nWant %v\nGot  %v", want, err)
		} else if buf.Len() != l {
			t.Errorf("Calling WriteParent() after calling WriteNode() ")
		}
	}
	// Calling WriteParent after Close() returns error
	{
		var buf bytes.Buffer
		w := NewWriter(&buf)
		w.closed = true
		l := buf.Len()
		want := fmt.Errorf("the writer is closed")
		if err := w.WriteParent(parentKey); !errEqual(want, err) {
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
		w.closed = true
		l := buf.Len()
		node := &testNode{key: key("k"), value: val("v")}
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
		err    error
	}{
		{
			name:   "Closing closed writer does nothing",
			closed: true,
			want:   "",
		},
		{
			name: "No nodes or parents have been written",
			want: "",
			err:  fmt.Errorf("must write atleast one node before closing"),
		},
		{
			name: "Only nodes have been written",
			node: true,
			want: "}",
		},
		{
			name:   "Only parent has been written",
			parent: true,
			want:   "",
			err:    fmt.Errorf("must write atleast one node before closing"),
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
		if err := w.Close(); !errEqual(test.err, err) {
			t.Errorf("%s: Close() returned wrong error\nWant %v\nGot  %v", test.name, test.err, err)
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
		nodes  []*testNode
		want   string
	}{
		{
			name: "Normal nodes slice without parent",
			nodes: []*testNode{
				{key: key("a"), nodes: []*testNode{
					{key: key("b"), value: val("v1")},
					{key: key("c"), nodes: []*testNode{
						{key: key("d"), value: val("v2")},
						{key: key("e"), value: val("v3")},
					}},
				}},
				{key: key("f"), nodes: []*testNode{
					{key: key("g"), value: val("v4")}}},
				{key: key("h"), nodes: []*testNode{
					{key: key("i"), nodes: []*testNode{
						{key: key("j"), value: val("v5")},
					}},
				}},
			},
			want: `{"a":{"b":"v1","c":{"d":"v2","e":"v3"}},"f":{"g":"v4"},"h":{"i":{"j":"v5"}}}`,
		},
		{
			name:   "Normal nodes slice with parent",
			parent: "root",
			nodes: []*testNode{
				{key: key("a"), nodes: []*testNode{
					{key: key("b"), value: val("v1")},
					{key: key("c"), nodes: []*testNode{
						{key: key("d"), value: val("v2")},
						{key: key("e"), value: val("v3")},
					}},
				}},
				{key: key("f"), nodes: []*testNode{
					{key: key("g"), value: val("v4")}}},
				{key: key("h"), nodes: []*testNode{
					{key: key("i"), nodes: []*testNode{
						{key: key("j"), value: val("v5")},
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
			if err := w.WriteParent(key(test.parent)); err != nil {
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
			parent := &testNode{key: key(test.parent), nodes: test.nodes}
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
				gotErr = w.WriteParent(key(test.parent))
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
				leaf.value = valErr("", wantErr, nil)
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

// ========== Benchmarking ==========

func benchmarkWriter(n int, b *testing.B) {
	node := getTestNode(n, n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWriter(discardWriter{})
		if err := w.WriteParent(node.key); err != nil {
			b.Fatalf("WriteParent() returned error: %v", err)
		}
		for _, node := range node.nodes {
			if err := w.WriteNode(node); err != nil {
				b.Fatalf("WriteNode() returned error: %v", err)
			}
		}
		if err := w.Close(); err != nil {
			b.Fatalf("Close() returned error: %v", err)
		}
	}
}

func BenchmarkWriter1(b *testing.B) { benchmarkWriter(1, b) }
func BenchmarkWriter2(b *testing.B) { benchmarkWriter(2, b) }
func BenchmarkWriter3(b *testing.B) { benchmarkWriter(3, b) }
func BenchmarkWriter4(b *testing.B) { benchmarkWriter(4, b) }
func BenchmarkWriter5(b *testing.B) { benchmarkWriter(5, b) }

// ========== Utility ==========

// memWriter writes to an internal buffer
type memWriter struct {
	bs []byte
}

func (mw *memWriter) Write(p []byte) (int, error) {
	mw.bs = append(mw.bs, p...)
	return len(p), nil
}

func findLeafNode(nodes []*testNode) *testNode {
	for _, node := range nodes {
		if len(node.nodes) == 0 {
			return node
		}
	}
	for _, node := range nodes {
		if leaf := findLeafNode(node.nodes); leaf != nil {
			return leaf
		}
	}
	return nil
}
