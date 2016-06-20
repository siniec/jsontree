package jsontree

import (
	"bytes"
	"fmt"
	"io"
)

type Node struct {
	Key   string
	Value string
	Nodes []*Node
}

func (node *Node) serializeTo(w io.Writer, wrapped bool) error {
	if wrapped {
		if _, err := w.Write([]byte{'{'}); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte(`"` + node.Key + `":`)); err != nil {
		return err
	}
	if node.Value != "" {
		if _, err := w.Write([]byte(`"` + node.Value + `"`)); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte{'{'}); err != nil {
			return err
		}
		n := len(node.Nodes)
		for i, child := range node.Nodes {
			if err := child.serializeTo(w, false); err != nil {
				return err
			}
			if hasMoreChildren := i < n-1; hasMoreChildren {
				if _, err := w.Write([]byte{','}); err != nil {
					return err
				}
			}
		}
		if _, err := w.Write([]byte{'}'}); err != nil {
			return err
		}
	}
	if wrapped {
		if _, err := w.Write([]byte{'}'}); err != nil {
			return err
		}
	}
	return nil
}

func (node *Node) SerializeTo(w io.Writer) error {
	return node.serializeTo(w, true)
}

func (node *Node) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	if err := node.SerializeTo(&buf); err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}

func (node *Node) DeserializeFrom(r io.Reader) error {
	node.Key = ""
	node.Nodes = nil
	node.Value = ""
	p := newParser(r)
	for p.Scan() {
		path, value := p.Data()
		node.Key = path[0]
		if len(path) == 1 {
			if node.Value != "" {
				return fmt.Errorf("invalid json. Expected 1 root node")
			}
			node.Value = value
		} else {
			node.getOrAdd(path[1:]...).Value = value
		}
	}
	return p.Err()
}

func (node *Node) Deserialize(b []byte) error {
	buf := bytes.NewBuffer(b)
	return node.DeserializeFrom(buf)
}
