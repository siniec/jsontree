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
	if node.Value != nil {
		if err := serializeValue(node.Value, w); err != nil {
			return err
		}
	} else {
		if err := serializeNodes(node.Nodes, w); err != nil {
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

func serializeValue(value Value, w io.Writer) error {
	if _, err := w.Write([]byte{'"'}); err != nil {
		return err
	}
	if b, err := value.Serialize(); err != nil {
		return err
	} else if _, err = w.Write(b); err != nil {
		return err
	}
	if _, err := w.Write([]byte{'"'}); err != nil {
		return err
	}
	return nil
}

func serializeNodes(nodes []*Node, w io.Writer) error {
	if _, err := w.Write([]byte{'{'}); err != nil {
		return err
	}
	n := len(nodes)
	for i, child := range nodes {
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
	return nil
}
