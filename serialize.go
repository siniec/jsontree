package jsontree

import (
	"bufio"
	"fmt"
	"io"
)

type ByteWriter interface {
	io.Writer
	io.ByteWriter
}

func SerializeNode(node Node, w io.Writer) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	var bw ByteWriter
	if _bw, ok := w.(ByteWriter); ok {
		bw = _bw
	} else {
		bw = bufio.NewWriter(w)
	}
	if err := bw.WriteByte('{'); err != nil {
		return err
	}
	if err := serializeNode(node, bw); err != nil {
		return err
	}
	return bw.WriteByte('}')
}

func serializeNode(node Node, w ByteWriter) error {
	if err := w.WriteByte('"'); err != nil {
		return err
	}
	if _, err := w.Write(node.Key()); err != nil {
		return err
	}
	if _, err := w.Write(jsonBytes[1:]); err != nil { // write ":
		return err
	}
	if nodes := node.Nodes(); len(nodes) > 0 {
		if err := serializeNodes(nodes, w); err != nil {
			return err
		}
	} else if value := node.Value(); value != nil {
		if err := serializeValue(value, w); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid node: len(node.Nodes()) == 0 and node.Value() == nil")
	}
	return nil
}

func serializeValue(value Value, w ByteWriter) error {
	if err := w.WriteByte('"'); err != nil {
		return err
	}
	if b, err := value.Serialize(); err != nil {
		return err
	} else if _, err = w.Write(b); err != nil {
		return err
	}
	if err := w.WriteByte('"'); err != nil {
		return err
	}
	return nil
}

func serializeNodes(nodes []Node, w ByteWriter) error {
	if err := w.WriteByte('{'); err != nil {
		return err
	}
	n := len(nodes)
	for i, node := range nodes {
		if node == nil {
			return fmt.Errorf("invalid node: node.Nodes() contained nil")
		}
		if err := serializeNode(node, w); err != nil {
			return err
		}
		if hasMoreChildren := i < n-1; hasMoreChildren {
			if err := w.WriteByte(','); err != nil {
				return err
			}
		}
	}
	if err := w.WriteByte('}'); err != nil {
		return err
	}
	return nil
}
