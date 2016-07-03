package jsontree

import (
	"bytes"
)

type Value interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

type Node interface {
	Key() []byte
	SetKey(key []byte)
	Value() Value
	Nodes() []Node
	AddNode(key []byte) Node
}

func getNode(node Node, path ...[]byte) Node {
	// no need to check len(path). get is only called by getOrAdd, which does that already
	key := path[0]
	for _, child := range node.Nodes() {
		if keyEqual(child.Key(), key) {
			if len(path) == 1 {
				return child
			} else {
				return getNode(child, path[1:]...)
			}
		}
	}
	return nil
}

func getOrAddNode(node Node, path ...[]byte) Node {
	if len(path) == 0 {
		return nil
	}
	key := path[0]
	n := getNode(node, key)
	if n == nil {
		n = node.AddNode(key)
	}
	if len(path) == 1 {
		return n
	} else {
		return getOrAddNode(n, path[1:]...)
	}
}

func keyEqual(key1, key2 []byte) bool {
	return bytes.Equal(key1, key2)
}
