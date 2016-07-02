package jsontree

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestSerializeNode(t *testing.T) {
	node := &Node{
		Key: key("data"),
		Nodes: []*Node{
			{
				Key: key("1"),
				Nodes: []*Node{
					{Key: key("a"), Value: val("v1")},
					{Key: key("b"), Nodes: []*Node{{Key: key("i"), Value: val("v2")}}},
				},
			},
			{
				Key: key("2"),
				Nodes: []*Node{
					{Key: key("a"), Value: val("v3")},
					{Key: key("b"), Nodes: []*Node{{Key: key("i"), Value: val("v4")}}},
				},
			},
		},
	}
	const want = `{"data":{"1":{"a":"v1","b":{"i":"v2"}},"2":{"a":"v3","b":{"i":"v4"}}}}`

	// Test successful serialization
	{
		var buf bytes.Buffer
		if err := SerializeNode(node, &buf); err != nil {
			t.Errorf("Unexpected error: %v", err)
		} else if got := buf.String(); want != got {
			t.Errorf("Wrong JSON written.\nWant %s\nGot  %s", want, got)
		}
	}

	// Test error with writer
	for i := 0; i <= len(want); i++ {
		wantErr := fmt.Errorf("Test err")
		w := &errWriter{
			errIndex: i,
			err:      wantErr,
		}
		gotErr := SerializeNode(node, w)
		if !errEqual(wantErr, gotErr) {
			t.Errorf("Wrong error returned\nWant %v\nGot  %v", wantErr, gotErr)
		}
	}

	// Test serialization error
	{
		// Set one of the nodes' values to return an error when serializing
		wantErr := fmt.Errorf("Serialize test err")
		node.Nodes[0].Nodes[0].Value = valErr("", wantErr, nil)
		gotErr := SerializeNode(node, ioutil.Discard)
		if !errEqual(wantErr, gotErr) {
			t.Errorf("Wrong error returned\nWant %v\nGot  %v", wantErr, gotErr)
		}
	}
}

// ========== Benchmarking ==========

func getTestNode(width, depth int) *Node {
	var fn func(i, width, depth int) *Node
	fn = func(i, width, depth int) *Node {
		node := &Node{
			Key: key(fmt.Sprintf("%d_%d", depth, i)),
		}
		if depth == 0 {
			node.Value = val("NodeVal")
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
	node := getTestNode(n-1, n-1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := SerializeNode(node, discardWriter{}); err != nil {
			b.Fatalf("Error: %v", err)
		}
	}
}

func BenchmarkNodeSerialization1(b *testing.B) { benchmarkNodeSerialization(1, b) }
func BenchmarkNodeSerialization2(b *testing.B) { benchmarkNodeSerialization(2, b) }
func BenchmarkNodeSerialization3(b *testing.B) { benchmarkNodeSerialization(3, b) }
func BenchmarkNodeSerialization4(b *testing.B) { benchmarkNodeSerialization(4, b) }
func BenchmarkNodeSerialization5(b *testing.B) { benchmarkNodeSerialization(5, b) }

// =========== Utility =============

type testValue struct {
	b            []byte
	serializeErr error
}

func (v *testValue) Serialize() ([]byte, error) {
	if v.serializeErr != nil {
		return nil, v.serializeErr
	}
	return v.b, nil
}

func (v *testValue) Deserialize(b []byte) error {
	v.b = b
	return nil
}

func (v *testValue) Equal(other *testValue) bool {
	return bytes.Equal(v.b, other.b)
}

func val(s string) *testValue {
	return valErr(s, nil, nil)
}

func valErr(s string, serializeErr, deserializeErr error) *testValue {
	return &testValue{
		b:            []byte(s),
		serializeErr: serializeErr,
	}
}

// errWriter acts like it writes the bytes passed to Write() and WriteByte() successfully
// until a certain number of bytes has been written.
type errWriter struct {
	errIndex int   // number of bytes to write successfully, without returning error
	err      error // error to return
	count    int   // current number of bytes written
}

func (w *errWriter) Write(b []byte) (int, error) {
	w.count += len(b)
	if w.count >= w.errIndex {
		return 0, w.err
	}
	return len(b), nil
}

func (w *errWriter) WriteByte(b byte) error {
	_, err := w.Write([]byte{b})
	return err
}

// discardWriter implements io.Writer, io.ByteWriter and discards anything written.
// It is used to prevent any overhead during benchmarking.
type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (discardWriter) WriteByte(b byte) error {
	return nil
}
