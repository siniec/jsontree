package jsontree

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Scanner struct {
	r    *bufio.Reader
	node *Node
	err  error

	hasReadStart bool
	eof          bool
}

func NewScanner(r io.Reader) *Scanner {
	s := new(Scanner)
	if bw, ok := r.(*bufio.Reader); ok {
		s.r = bw
	} else {
		s.r = bufio.NewReader(r)
	}
	return s
}

func (scanner *Scanner) Scan() bool {
	if scanner.err != nil || scanner.eof {
		return false
	}

	r := scanner.r

	if !scanner.hasReadStart {
		scanner.hasReadStart = true
		if b, err := r.ReadByte(); err != nil {
			scanner.err = err
			return false
		} else if b != '{' {
			scanner.err = fmt.Errorf("Invalid input: did not start with '{'. Found '%s'", string(b))
			return false
		}
	}
	// Start reading the next node
	// Strategy:
	//  Read until the next "}". Then count the number of "{" we've gone over, nOpens.
	// If nOpens == 0, we're at the end (there was one closing bracket)
	// If nOpens > 0, then read the equal number of "}" to close the object.

	if obj, err := r.ReadString('}'); err != nil {
		scanner.err = err
		return false
	} else {
		if obj == "}" { // this happens when the input is just "{}"
			scanner.eof = true
			return false
		}
		var b bytes.Buffer // b holds the current object we're reading
		b.WriteByte('{')   // the string we're reading is in the form "key": {...}. We must wrap it in braces to be valid for mashalling
		b.WriteString(obj)
		nOpening := strings.Count(obj, "{")
		nMissingClosing := nOpening - 1
		for i := 0; i < nMissingClosing; i++ {
			if part, err := r.ReadString('}'); err != nil {
				if err == io.EOF {
					scanner.err = fmt.Errorf("Invalid input: JSON object is incomplete (EOF reached)")
				} else {
					scanner.err = err
				}
				return false
			} else {
				nMissingClosing += strings.Count(part, "{")
				b.WriteString(part)
			}
		}
		// Check if it's the end of the object. We expect a comma after the current node. If it's not present, check that it
		if bs, err := r.Peek(1); err != nil {
			if err == io.EOF {
				scanner.err = fmt.Errorf("Invalid input: JSON object is incomplete (EOF reached)")
			} else {
				scanner.err = err
			}
			return false
		} else {
			switch bs[0] {
			case ',':
				// then there's another key:object after this one. Read that one comma to prepare for the next Scan()
				if _, err := r.ReadByte(); err != nil {
					scanner.err = err
					return false
				}
			case '}':
				// Then it's the end of the root node. Note: don't return here. Just mark EOF for the next call to Scan(),
				// as we're currently reading a node
				scanner.eof = true
			default:
				scanner.err = fmt.Errorf("Invalid input: unknown token encountered: '%v'. Expected ',' or '}'", string(bs[0]))
				return false
			}
		}

		b.WriteByte('}')
		node := new(Node)
		if err := node.DeserializeFrom(&b); err != nil {
			scanner.err = fmt.Errorf("Error deserializing node: %v", err)
			return false
		}
		scanner.node = node
		return true
	}
}

func (scanner *Scanner) Node() *Node {
	return scanner.node
}

func (scanner *Scanner) Err() error {
	return scanner.err
}
