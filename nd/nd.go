package nd

type Node struct {
	next  *Node
	prev  *Node
	Value interface{}
}

func New(value interface{}) *Node {
	node := &Node{Value: value}
	node.next = node
	node.prev = node
	return node
}

func (node *Node) Len() int {
	var length = 1
	for n := node.next; n != node; n = n.next {
		length++
	}
	return length
}

func (node *Node) Next() *Node {
	return node.next
}

func (node *Node) Prev() *Node {
	return node.prev
}

func (a *Node) Link(b *Node) {
	n := a.next
	p := b.prev

	a.next = b
	b.prev = a
	n.prev = p
	p.next = n
}

func (node *Node) Unlink() {
	node.prev.Link(node.next)
}

func (node *Node) Select(n int) *Node {
	switch {
	case n < 0:
		for ; n < 0; n++ {
			node = node.prev
		}
	case n > 0:
		for ; n > 0; n-- {
			node = node.next
		}
	}
	return node
}

func (node *Node) Swap(other *Node) {
	other.Value, node.Value = node.Value, other.Value
}

func (node *Node) All() []*Node {
	nodes := []*Node{node}
	for n := node.next; n != node; n = n.next {
		nodes = append(nodes, n)
	}
	return nodes
}

func (node *Node) Each() chan *Node {
	ch := make(chan *Node)
	go func() {
		ch <- node
		for n := node.next; n != node; n = n.next {
			ch <- n
		}
		close(ch)
	}()
	return ch
}
