package jsontree

import (
	"encoding/json"
	"fmt"
	"io"
)

type parser struct {
	dec   *json.Decoder
	nodes stack
	next  parseFn
}

func (p *parser) node() *Node {
	return p.nodes.Peek(0)
}

func (p *parser) pushChild() *Node {
	// fmt.Println("pushChild()")
	node := new(Node)
	if curr := p.nodes.Peek(0); curr != nil {
		curr.Nodes = append(curr.Nodes, node)
		// fmt.Printf("\tAppending as child (%v) (%v)\n", curr, p.nodes[0])
	}
	p.nodes.Push(node)
	return node
}

func (p *parser) pushSibling() *Node {
	// fmt.Println("pushSibling()")
	node := new(Node)
	if curr := p.nodes.Peek(1); curr != nil {
		curr.Nodes = append(curr.Nodes, node)
		// fmt.Printf("\tAppending as sibling (%v) (%v)\n", curr, p.nodes[0])
	}
	p.nodes.Push(node)
	return node
}

func (p *parser) popNode() {
	p.nodes.Pop()
}

func newParser(r io.Reader) *parser {
	p := &parser{
		dec: json.NewDecoder(r),
	}
	p.next = p.readOpenBracket
	return p
}

type nextFn func(node *Node, t json.Token) (nextFn, error)
type parseFn func(t json.Token) error

func (p *parser) Parse(node *Node) error {
	// read the first token.
	if t, err := p.dec.Token(); err != nil {
		return err
	} else if err := p.readDelim(t, '{'); err != nil {
		return err
	}
	p.nodes.Push(node)
	for {
		t, err := p.dec.Token()
		// fmt.Println("Token:", t)
		if err == io.EOF {
			p.next = func(_ json.Token) error { return fmt.Errorf("parser has already parsed") }
			break
		} else if err != nil {
			return err
		} else if p.node() == nil {
			return fmt.Errorf("expected JSON input to end. Found '%s'", t)
		} else if err = p.next(t); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) readDelim(t json.Token, d json.Delim) error {
	if t, ok := t.(json.Delim); !ok || t != d {
		return fmt.Errorf("invalid JSON '%s' (%T) looking for '%s' (json.Delim)", t, t, d)
	}
	p.next = p.readKey
	return nil
}

func (p *parser) readOpenBracket(t json.Token) error {
	if err := p.readDelim(t, '{'); err != nil {
		return err
	}
	// fmt.Println("readOpenBracket()")
	p.next = p.readKey
	p.pushChild()
	return nil
}
func (p *parser) readCloseBracket(t json.Token) error {
	if err := p.readDelim(t, '}'); err != nil {
		return err
	}
	p.nodes.Pop()
	// if len(p.nodes) > 0 {
	// 	fmt.Printf("readCloseBracket() (current node: %v) (top: %v)\n", p.nodes.Peek(0), p.nodes[0])
	// }
	p.next = merge(p.readCloseBracket, p.readKey)
	return nil
}

func (p *parser) readKey(t json.Token) error {
	if key, ok := t.(string); !ok {
		return fmt.Errorf("invalid JSON '%s' (%T) looking for key", t, t)
	} else {
		// fmt.Println("readKey()")
		p.node().Key = key
		// fmt.Println("\tRead key successfully", key)
		p.next = merge(p.readValue, p.readOpenBracket)
		return nil
	}
}

func (p *parser) readSiblingKey(t json.Token) error {
	if key, ok := t.(string); !ok {
		return fmt.Errorf("invalid JSON '%s' (%T) looking for key", t, t)
	} else {
		// fmt.Println("readSiblingKey()", key)
		p.pushSibling()
		p.node().Key = key
		// fmt.Println("\tRead key successfully", key)
		p.next = merge(p.readValue, p.readOpenBracket)
		return nil
	}
}

func (p *parser) readValue(t json.Token) error {
	if val, ok := t.(string); !ok {
		return fmt.Errorf("invalid JSON '%s' (%T) looking for value string", t, t)
	} else {
		// fmt.Println("\treadValue()", val)
		p.node().Value = val
		p.next = merge(p.readCloseBracket, p.readSiblingKey)
		return nil
	}
}

func merge(fns ...parseFn) parseFn {
	return func(t json.Token) error {
		var err error
		for _, fn := range fns {
			if err = fn(t); err == nil {
				return nil
			}
		}
		return err
	}
}

func (node *Node) DeserializeFrom(r io.Reader) error {
	p := newParser(r)
	return p.Parse(node)
}

type stack []*Node

func (s *stack) Push(v *Node) {
	*s = append(*s, v)
}

func (s *stack) Pop() *Node {
	if n := len(*s); n == 0 {
		return nil
	} else {
		node := (*s)[n-1]
		*s = (*s)[:n-1]
		return node
	}
}

func (s stack) Peek(level int) *Node {
	if n := len(s); n <= level {
		return nil
	} else {
		return s[n-1-level]
	}
}
