package jsontree

import (
	"bufio"
	"fmt"
	"io"
)

type GetValueFn func() Value

func DeserializeNode(r io.Reader, getValFn GetValueFn) (*Node, error) {
	node := new(Node)
	p := newParser(r, getValFn)
	for p.Scan() {
		path, value := p.Data()
		node.Key = path[0]
		if len(path) == 1 {
			if node.Value != nil {
				return node, fmt.Errorf("invalid json. Expected 1 root node")
			}
			node.Value = value
		} else {
			node.getOrAdd(path[1:]...).Value = value
		}
	}
	return node, p.Err()
}

type readFn func() (next readFn, err error)

type parser struct {
	r     *bufio.Reader
	valFn GetValueFn
	next  readFn
	err   error
	mode  int
	path  stack
	value Value
	eof   bool
}

func newParser(r io.Reader, getValFn GetValueFn) *parser {
	p := &parser{valFn: getValFn}
	if br, ok := r.(*bufio.Reader); ok {
		p.r = br
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

func (p *parser) Data() (path []string, value Value) {
	path = p.path
	value = p.value
	return path, value
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
		return nil, fmt.Errorf("Read '%s', want '%s'", string(bGot), string(bWant))
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
			return nil, fmt.Errorf(`Read '%s', want '{' or '"'`, string(bs))
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
		p.path.Push(string(bs))
	}
	// Following the key should be a column
	if _, err := p.readByte(':', nil); err != nil {
		return nil, err
	}
	// And following the column is eiter a sub node or a value
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
			return nil, fmt.Errorf(`Read %s, want '{' or '"'`, string(bs))
		}
	}
}

func (p *parser) readQuotedValue() (readFn, error) {
	if bs, err := p.readQuotedString(); err != nil {
		return nil, err
	} else {
		p.value = p.valFn()
		p.err = p.value.Deserialize(bs)
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
			return nil, fmt.Errorf(`Read '%s', want '}' or ','`, string(bs))
		}
	}
}

func (p *parser) readComma() (readFn, error) {
	p.path.Pop()
	return p.readByte(',', p.readQuotedKey)
}

type stack []string

func (s *stack) Push(v string) {
	*s = append(*s, v)
}

func (s *stack) Pop() {
	if n := len(*s); n > 0 {
		*s = (*s)[:n-1]
	}
}
