package jsontree

import (
	"bytes"
)

type Value interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

type Node struct {
	Key   []byte
	Value Value
	Nodes []*Node
}

func (node *Node) get(path ...[]byte) *Node {
	// no need to check len(path). get is only called by getOrAdd, which does that already
	key := path[0]
	for _, child := range node.Nodes {
		if keyEqual(child.Key, key) {
			if len(path) == 1 {
				return child
			} else {
				return child.get(path[1:]...)
			}
		}
	}
	return nil
}

func (node *Node) getOrAdd(path ...[]byte) *Node {
	if len(path) == 0 {
		return nil
	}
	key := path[0]
	n := node.get(key)
	if n == nil {
		n = &Node{Key: key}
		node.Nodes = append(node.Nodes, n)
	}
	if len(path) == 1 {
		return n
	} else {
		return n.getOrAdd(path[1:]...)
	}
}

func keyEqual(key1, key2 []byte) bool {
	return bytes.Equal(key1, key2)
}
