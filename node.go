package jsontree

import (
	"bytes"
	"encoding/json"
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

// func (node *Node) DeserializeFrom(r io.Reader) error {
// 	var buf bytes.Buffer
// 	if _, err := buf.ReadFrom(r); err != nil {
// 		return err
// 	}
// 	return node.Deserialize(buf.Bytes())
// }

func (node *Node) Deserialize(b []byte) error {
	var parse func(node *Node, obj interface{}) error
	parse = func(node *Node, obj interface{}) error {
		switch obj.(type) {
		case string:
			node.Value = obj.(string)
		case map[string]interface{}:
			for key, val := range obj.(map[string]interface{}) {
				n := &Node{Key: key}
				node.Nodes = append(node.Nodes, n)
				parse(n, val)
			}
		default:
			return fmt.Errorf("Invalid json: unknown type parsed: %T", obj)
		}
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	} else {
		m := v.(map[string]interface{})
		if len(m) != 1 {
			return fmt.Errorf("Invalid json. Expected 1 root note. got %d", len(m))
		}
		for key, val := range m {
			node.Key = key
			parse(node, val)
		}
	}
	return nil
}
