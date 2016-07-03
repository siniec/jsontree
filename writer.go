package jsontree

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

var jsonBytes = []byte{'{', '"', ':'}

type Writer struct {
	w                ByteWriter
	hasWrittenNode   bool
	hasWrittenParent bool
	closed           bool
}

func NewWriter(w io.Writer) *Writer {
	writer := new(Writer)
	if bw, ok := w.(ByteWriter); ok {
		writer.w = bw
	} else {
		writer.w = bufio.NewWriter(w)
	}
	return writer
}

func (writer *Writer) WriteParent(key []byte) error {
	if writer.closed {
		return errors.New("the writer is closed")
	}
	if writer.hasWrittenNode {
		return errors.New("WriteParent() must be called before any call to WriteNode()")
	}
	if writer.hasWrittenParent {
		return errors.New("WriteParent() has already been called")
	}
	if _, err := writer.w.Write(jsonBytes[:2]); err != nil { // write {"
		return err
	}
	if _, err := writer.w.Write(key); err != nil {
		return err
	}
	if _, err := writer.w.Write(jsonBytes[1:]); err != nil { // write ":
		return err
	}
	writer.hasWrittenParent = true
	return nil
}

func (writer *Writer) WriteNode(node Node) error {
	if writer.closed {
		return errors.New("the writer is closed")
	}
	w := writer.w
	if !writer.hasWrittenNode {
		if err := w.WriteByte('{'); err != nil {
			return err
		}
		writer.hasWrittenNode = true
	} else {
		if err := w.WriteByte(','); err != nil {
			return err
		}
	}
	return serializeNode(node, w)
}

func (writer *Writer) Close() error {
	if writer.closed {
		return nil
	}
	if !writer.hasWrittenNode {
		return fmt.Errorf("must write atleast one node before closing")
	}
	w := writer.w
	if err := w.WriteByte('}'); err != nil {
		return err
	}
	if writer.hasWrittenParent {
		if err := w.WriteByte('}'); err != nil {
			return err
		}
	}
	writer.closed = true
	return nil
}
