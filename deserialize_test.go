package jsontree

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestDeserializeError(t *testing.T) {
	err := &DeserializeError{
		Got:  '?',
		Want: []byte{'a', 'b', 'c'},
	}
	want := "Read '?', expected 'a' or 'b' or 'c'"
	if err.Error() != want {
		t.Errorf("Error() = %v, want %v", err.Error(), want)
	}
}

func TestParserScan(t *testing.T) {
	// Scan() returns false if error
	{
		p := parser{err: fmt.Errorf("Err")}
		if p.Scan() {
			t.Errorf("Scan() returned true when parser has error. Should return false.")
		}
	}
}

func TestDeserializeNode(t *testing.T) {
	tests := []struct {
		in    string
		weird bool // signal not to try creating invalid JSON out of this test input
		want  *Node
		err   error
	}{
		{
			// Top level is leaf
			in:   `{"a":"b"}`,
			want: &Node{Key: key("a"), Value: val("b")},
		},
		{
			// Sibling leaf nodes
			in: `{"root":{"a":"b","c":"d"}}`,
			want: &Node{Key: key("root"), Nodes: []*Node{
				{Key: key("a"), Value: val("b")},
				{Key: key("c"), Value: val("d")},
			}},
		},
		{
			// Sibling non-leaf nodes
			in: `{"root":{"a":{"a1":"v1"},"b":{"b1":"v2"}}}`,
			want: &Node{Key: key("root"), Nodes: []*Node{
				{Key: key("a"), Nodes: []*Node{
					{Key: key("a1"), Value: val("v1")},
				}},
				{Key: key("b"), Nodes: []*Node{
					{Key: key("b1"), Value: val("v2")},
				}},
			}},
		},
		{
			// Leaf nodes on different levels
			in: `{"root":{"a":"v1","b":{"b1":{"b11":"v3"},"b2":"v2"}}}`,
			want: &Node{Key: key("root"), Nodes: []*Node{
				{Key: key("a"), Value: val("v1")},
				{Key: key("b"), Nodes: []*Node{
					{Key: key("b1"), Nodes: []*Node{
						{Key: key("b11"), Value: val("v3")},
					}},
					{Key: key("b2"), Value: val("v2")},
				}},
			}},
		},
		{
			// Nodes are ordered non-alphabetically
			in: `{"root":{"b":"3","c":"1","a":"2"}}`,
			want: &Node{Key: key("root"), Nodes: []*Node{
				{Key: key("b"), Value: val("3")},
				{Key: key("c"), Value: val("1")},
				{Key: key("a"), Value: val("2")},
			}},
		},
		// Weird but valid input
		{
			in:    `{"ro\"ot":{"{a}":"\"hello\"","b}":"\\backslash\nnewline"}}`,
			weird: true,
			want: &Node{Key: key(`ro\"ot`), Nodes: []*Node{
				{Key: key(`{a}`), Value: val(`\"hello\"`)},
				{Key: key(`b}`), Value: val(`\\backslash\nnewline`)},
			}},
		},
		// Handling invalid input. See also section Test unexpected tokens (invalid JSON) below
		// -- JSON syntax error
		{
			in:  `{"a":"b"},`, // extra, invalid comma
			err: fmt.Errorf("expected end of input. Got ','"),
		},
		{
			in:  `{"a":"b"`, // json ends abruptly
			err: fmt.Errorf("reader returned io.EOF before expected"),
		},
		{
			in:  `{"a`, // key is never closed
			err: fmt.Errorf("reader returned io.EOF before expected"),
		},
		// -- Semantic error
		{
			in:  `{"a":"b","c":"d"}`,
			err: fmt.Errorf("invalid json. Expected 1 root node"),
		},
	}
	getValFn := func() Value { return new(testValue) }
	for _, test := range tests {
		r := bytes.NewReader([]byte(test.in))
		node, err := DeserializeNode(r, getValFn)
		if test.err != nil {
			if !errEqual(test.err, err) {
				t.Errorf("%s\nWrong error.\nWant %v\nGot  %v", test.in, test.err, err)
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

		// Test reader error
		{
			// Error after x bytes read
			for i := 0; i <= len(test.in); i++ {
				wantErr := fmt.Errorf("Reader test error")
				r := &readPeeker{
					r: &errReader{
						br:       bytes.NewReader([]byte(test.in)),
						errIndex: i,
						err:      wantErr,
					},
				}
				_, gotErr := DeserializeNode(r, getValFn)
				if !errEqual(wantErr, gotErr) {
					t.Errorf("%s (errReader(%d)\nWrong error.\nWant %v\nGot  %v", test.in, i, wantErr, gotErr)
				}
			}
			// Error on ReadByte() after successful call to Peek()
			for i := 1; i <= len(test.in); i++ {
				wantErr := fmt.Errorf("ReadByte() test error")
				r := &readPeeker{
					r:        bytes.NewReader([]byte(test.in)),
					readErr:  wantErr,
					errIndex: i,
				}
				_, gotErr := DeserializeNode(r, getValFn)
				if r.peekCount < i {
					// we've reached the max number of Peek() calls for this input
					break
				}
				if !errEqual(wantErr, gotErr) {
					t.Errorf("%s (ReadByte() error)\nWrong error.\nWant %v\nGot  %v", test.in, wantErr, gotErr)
				}
			}
		}
		// Test unexpected tokens (invalid JSON)
		if test.err == nil && !test.weird {
			validJSON := test.in
			var prev rune
			for i, char := range validJSON {
				switch char {
				case '{', '}', '[', ']', ':', '"', '-', ',':
					if char == '"' && isAlphanumeric(prev) {
						continue // its the end of a key or value string. This is handled in another test.
					}
					// Replace the character with a ?. Eg for {"a":"b"}, we test [?, {?, {"a?, {"a"?, {"a":?, ...]
					invalidJSON := validJSON[:i] + "?"
					r := strings.NewReader(invalidJSON)
					_, err := DeserializeNode(r, getValFn)
					if _, ok := err.(*DeserializeError); !ok {
						t.Errorf(`DeserializeNode(%s): Wrong error type. Want DeserializeError, got %T ("%v")`, invalidJSON, err, err)
					}
				}
				prev = char
			}
		}
	}
}

// ========== Benchmarking ==========

func benchmarkNodeDeserialization(n int, b *testing.B) {
	node := getTestNode(n-1, n-1)
	var buf bytes.Buffer
	if err := SerializeNode(node, &buf); err != nil {
		panic(err)
	}
	getValFn := func() Value { return new(testValue) }
	r := new(bytes.Buffer)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Reset()
		if _, err := r.Write(buf.Bytes()); err != nil {
			b.Fatalf("r.Write() error: %v", err)
		}
		if _, err := DeserializeNode(r, getValFn); err != nil {
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
	if !keyEqual(got.Key, want.Key) {
		return false
	}
	if (got.Value == nil && want.Value != nil) || (got.Value != nil && want.Value == nil) {
		return false
	}
	if got.Value != nil && !got.Value.(*testValue).Equal(want.Value.(*testValue)) {
		return false
	}
	return nodesEqual(want.Nodes, got.Nodes)
}

func nodesEqual(want, got []*Node) bool {
	if len(got) != len(want) || (got == nil && want != nil) || (got != nil && want == nil) {
		return false
	}
	for i := range want {
		if !nodeEqual(want[i], got[i]) {
			return false
		}
	}
	return true
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

// errReader reads from the underlying reader until a certain number of
// bytes has been read.
type errReader struct {
	br       *bytes.Reader
	errIndex int   // number of bytes to read successfully, without returning error
	err      error // error to return
	count    int   // current number of bytes read
}

func (r *errReader) Read(b []byte) (int, error) {
	r.count += len(b)
	if r.hasError() {
		return 0, r.err
	}
	return r.br.Read(b)
}

func (r *errReader) hasError() bool {
	return r.count >= r.errIndex
}

// readPeeker is a custom implementation of the ReadSeeker inferface,
// to escape the buffering of bufio and control when errors are returned
// when using errReader as the underlying reader
type readPeeker struct {
	peek      []byte
	r         io.Reader
	readErr   error // the error to be returned by ReadByte()
	errIndex  int   // after what number of calls to Peek() should ReadByte() return an error
	peekCount int   // number of times Peek() has been called
}

func (rp *readPeeker) Read([]byte) (int, error) {
	panic("Read is not implemented")
}

func (rp *readPeeker) ReadByte() (b byte, err error) {
	defer func() { rp.peek = nil }()
	if rp.shouldReturnReadError() {
		return b, rp.readErr
	}
	if len(rp.peek) > 0 {
		return rp.peek[0], nil
	}
	p := make([]byte, 1)
	if _, err = rp.r.Read(p); err != nil {
		return b, err
	}
	return p[0], nil
}

func (rp *readPeeker) Peek(n int) (b []byte, err error) {
	b = make([]byte, n)
	_, err = rp.r.Read(b)
	rp.peek = b
	rp.peekCount += 1
	return b, err
}

func (rp *readPeeker) lastActionWasPeek() bool {
	return rp.peek != nil
}

func (rp *readPeeker) shouldReturnReadError() bool {
	return rp.readErr != nil && rp.lastActionWasPeek() && rp.peekCount >= rp.errIndex
}
