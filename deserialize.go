package jsontree

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type DeserializeError struct {
	Got  byte
	Want []byte
}

func (err *DeserializeError) Error() string {
	wants := make([]string, len(err.Want))
	for i, want := range err.Want {
		wants[i] = string(want)
	}
	wantStr := strings.Join(wants, "' or '")
	return fmt.Sprintf("Read '%s', expected '%s'", string(err.Got), wantStr)
}

func DeserializeNode(node Node, r io.Reader) error {
	p := newParser(r)
	isKeySet := false
	for p.Scan() {
		path, valBytes := p.Data()
		n := len(path)
		if n == 1 && isKeySet {
			return fmt.Errorf("invalid json. Expected 1 root node")
		}
		if !isKeySet {
			node.SetKey(path[0])
			isKeySet = true
		}
		var value Value
		if n == 1 {
			value = node.Value()
		} else {
			child := getOrAddNode(node, path[1:]...)
			value = child.Value()
		}
		if err := value.Deserialize(valBytes); err != nil {
			return err
		}
	}
	return p.Err()
}

type readFn func() (next readFn, err error)

type ReadPeeker interface {
	ReadByte() (byte, error)
	Peek(int) ([]byte, error)
}

type parser struct {
	r     ReadPeeker
	next  readFn
	err   error
	mode  int
	path  stack
	value []byte
	eof   bool
}

func newParser(r io.Reader) *parser {
	p := new(parser)
	if rp, ok := r.(ReadPeeker); ok {
		p.r = rp
	} else {
		p.r = bufio.NewReader(r)
	}
	p.next = p.readOpenBracket
	return p
}

func (p *parser) Scan() bool {
	if p.err != nil {
		return false
	}
	p.value = nil
	// Scan until we've hit a value
	for p.value == nil {
		p.next, p.err = p.next()
		if p.err != nil {
			return false
		}
	}
	return true
}

func (p *parser) Data() (path [][]byte, valueBytes []byte) {
	path = p.path
	valueBytes = p.value
	return path, valueBytes
}

func (p *parser) Err() error {
	if p.err == io.EOF {
		if p.eof {
			return nil
		} else {
			return fmt.Errorf("reader returned io.EOF before expected")
		}
	} else {
		return p.err
	}
}

func (p *parser) readByte(bWant byte, next readFn) (readFn, error) {
	if bGot, err := p.r.ReadByte(); err != nil {
		return nil, err
	} else if bGot != bWant {
		return nil, &DeserializeError{Got: bGot, Want: []byte{bWant}}
	} else {
		return next, nil
	}
}

func (p *parser) readOpenBracket() (readFn, error) {
	return p.readByte('{', p.readQuotedKey)
}

func (p *parser) readCloseBracket() (readFn, error) {
	if _, err := p.readByte('}', nil); err != nil {
		return nil, err
	}
	p.path.Pop()
	if len(p.path) == 0 {
		p.eof = true
		if b, err := p.r.ReadByte(); err == nil {
			return nil, fmt.Errorf("expected end of input. Got '%s'", string(b))
		} else {
			return nil, err
		}
	}
	if bs, err := p.r.Peek(1); err != nil {
		return nil, err
	} else {
		switch bs[0] {
		case '}':
			return p.readCloseBracket, nil
		case ',':
			return p.readComma, nil
		default:
			return nil, &DeserializeError{Got: bs[0], Want: []byte{'}', ','}}
		}
	}
}

func (p *parser) readQuotedString() ([]byte, error) {
	if _, err := p.readByte('"', nil); err != nil {
		return nil, err
	}
	var bs []byte
	for {
		b, err := p.r.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == '\\' { // escape
			bs = append(bs, b)
			if b, err := p.r.ReadByte(); err != nil {
				return nil, err
			} else {
				bs = append(bs, b)
			}
			continue
		}
		if b == '"' {
			break
		}
		bs = append(bs, b)
	}
	return bs, nil
}

func (p *parser) readQuotedKey() (readFn, error) {
	if bs, err := p.readQuotedString(); err != nil {
		return nil, err
	} else {
		p.path.Push(bs)
	}
	// Following the key should be a column
	if _, err := p.readByte(':', nil); err != nil {
		return nil, err
	}
	// And following the column is either a sub node or a value
	if bs, err := p.r.Peek(1); err != nil {
		return nil, err
	} else {
		switch bs[0] {
		case '{':
			// Consume the byte
			if _, err := p.r.ReadByte(); err != nil {
				return nil, err
			}
			return p.readQuotedKey, nil
		case '"':
			return p.readQuotedValue, nil
		default:
			return nil, &DeserializeError{Got: bs[0], Want: []byte{'{', '"'}}
		}
	}
}

func (p *parser) readQuotedValue() (readFn, error) {
	if bs, err := p.readQuotedString(); err != nil {
		return nil, err
	} else {
		p.value = bs
	}
	// Following the value is either a sibling node or a closing bracket
	if bs, err := p.r.Peek(1); err != nil {
		return nil, err
	} else {
		switch bs[0] {
		case '}':
			return p.readCloseBracket, nil
		case ',':
			return p.readComma, nil
		default:
			return nil, &DeserializeError{Got: bs[0], Want: []byte{'}', ','}}
		}
	}
}

func (p *parser) readComma() (readFn, error) {
	p.path.Pop()
	return p.readByte(',', p.readQuotedKey)
}

type stack [][]byte

func (s *stack) Push(v []byte) {
	*s = append(*s, v)
}

func (s *stack) Pop() {
	if n := len(*s); n > 0 {
		*s = (*s)[:n-1]
	}
}
