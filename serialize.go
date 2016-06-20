package jsontree

import (
	"io"
)

func SerializeNode(node *Node, w io.Writer) error {
	return serializeNode(node, w, true)
}

func serializeNode(node *Node, w io.Writer, wrapped bool) error {
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
			if err := serializeNode(child, w, false); err != nil {
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
