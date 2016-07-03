package jsontree

import (
	"bufio"
	"io"
)

type ByteWriter interface {
	io.Writer
	io.ByteWriter
}

func SerializeNode(node Node, w io.Writer) error {
	var bw ByteWriter
	if _bw, ok := w.(ByteWriter); ok {
		bw = _bw
	} else {
		bw = bufio.NewWriter(w)
	}
	return serializeNode(node, bw, true)
}

func serializeNode(node Node, w ByteWriter, wrapped bool) error {
	if wrapped {
		if _, err := w.Write(jsonBytes[:2]); err != nil {
			return err
		}
	} else {
		if err := w.WriteByte('"'); err != nil {
			return err
		}
	}
	if _, err := w.Write(node.Key()); err != nil {
		return err
	}
	if _, err := w.Write(jsonBytes[3:5]); err != nil {
		return err
	}
	if nodes := node.Nodes(); len(nodes) > 0 {
		if err := serializeNodes(nodes, w); err != nil {
			return err
		}
	} else if value := node.Value(); value != nil {
		if err := serializeValue(node.Value(), w); err != nil {
			return err
		}
		// } else {
		// 	return fmt.Errorf("len(node.Nodes()) == 0 and node.Value() == nil")
	}
	if wrapped {
		if err := w.WriteByte('}'); err != nil {
			return err
		}
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
	for i, child := range nodes {
		if err := serializeNode(child, w, false); err != nil {
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
