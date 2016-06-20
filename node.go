package jsontree

type Node struct {
	Key   string
	Value string
	Nodes []*Node
}

func (node *Node) get(path ...string) *Node {
	// no need to check len(path). get is only called by getOrAdd, which does that already
	key := path[0]
	for _, child := range node.Nodes {
		if child.Key == key {
			if len(path) == 1 {
				return child
			} else {
				return child.get(path[1:]...)
			}
		}
	}
	return nil
}

func (node *Node) getOrAdd(path ...string) *Node {
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
