package jsontree

import (
	"errors"
	"io"
)

type Writer struct {
	w                io.Writer
	key              string
	hasWrittenNode   bool
	hasWrittenParent bool
	closed           bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (writer *Writer) WriteParent(key string) error {
	if writer.closed {
		return errors.New("the writer is closed")
	}
	if writer.hasWrittenNode {
		return errors.New("WriteParent() must be called before any call to WriteNode()")
	}
	if writer.hasWrittenParent {
		return errors.New("WriteParent() has already been called")
	}
	if _, err := writer.w.Write([]byte(`{"` + key + `":`)); err != nil {
		return err
	}
	writer.hasWrittenParent = true
	return nil
}

func (writer *Writer) WriteNode(node *Node) error {
	if writer.closed {
		return errors.New("the writer is closed")
	}
	w := writer.w
	if !writer.hasWrittenNode {
		if _, err := w.Write([]byte{'{'}); err != nil {
			return err
		}
		writer.hasWrittenNode = true
	} else {
		if _, err := w.Write([]byte{','}); err != nil {
			return err
		}
	}
	return node.serializeTo(w, false)
}

func (writer *Writer) Close() error {
	if writer.closed {
		return nil
	}
	w := writer.w
	if !writer.hasWrittenNode {
		if _, err := w.Write([]byte{'{'}); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte{'}'}); err != nil {
		return err
	}
	if writer.hasWrittenParent {
		if _, err := w.Write([]byte{'}'}); err != nil {
			return err
		}
	}
	writer.closed = true
	return nil
}
