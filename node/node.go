package node

type Node struct {
	prev  *Node
	next  *Node
	Value interface{}
}

func (node *Node) Insert(prev, next *Node) {
	next.prev = node
	node.next = next
	node.prev = prev
	prev.next = node
}

func New(value interface{}) *Node {
	node := &Node{Value: value}

	node.next = node
	node.prev = node

	return node
}

func (node *Node) Link(next *Node) {
	next.prev = node
	node.next = next
}

func (node *Node) Unlink() {
	node.prev.Link(node.next)
}

func (node *Node) Next(head *Node) *Node {
	next := node.next

	if head != nil && next == head {
		return head.next
	}

	return next
}

func (node *Node) Prev(head *Node) *Node {
	prev := node.prev

	if head != nil && prev == head {
		return head.prev
	}

	return prev
}

func (node *Node) Select(head *Node, n int) *Node {
	switch {
	case n < 0:
		for ; n < 0; n++ {
			node = node.Prev(head)
		}
	case n > 0:
		for ; n > 0; n-- {
			node = node.Next(head)
		}
	}

	return node
}

func (node *Node) Shift(head *Node, n int) {
	switch {
	case n < 0:
		for ; n < 0; n++ {
			// TODO
		}
	case n > 0:
		for ; n > 0; n-- {
			// TODO
		}
	}
}

func (node *Node) Swap(other *Node) {
	other.Value, node.Value = node.Value, other.Value
}

// Head

func (head *Node) Empty() bool {
	return head.next == head
}

func (head *Node) First() *Node {
	return head.next
}

func (head *Node) Append(node *Node) {
	node.Insert(head.prev, head)
}

func (head *Node) Len() (length int) {
	for n := head.next; n != head; n = n.next {
		length++
	}

	return
}

func (head *Node) All() []*Node {
	var nodes []*Node

	for n := head.next; n != head; n = n.next {
		nodes = append(nodes, n)
	}

	return nodes
}
