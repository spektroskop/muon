//go:generate genny -in=$GOFILE -out=window_list.go gen "Item=*Window"
//go:generate genny -in=$GOFILE -out=monitor_list.go gen "Item=*Monitor"
//go:generate genny -in=$GOFILE -out=layout_list.go gen "Item=Layout"

package muon

import (
	"fmt"
	"math"
	"strings"

	"github.com/dimchansky/genny/generic"
)

type Item generic.Type

type ItemNode struct {
	prev, next *ItemNode
	list       *ItemList
	Data       Item
}

type ItemList struct {
	focused *ItemNode
	root    ItemNode
	count   int
}

func (l *ItemList) FocusIsFront() bool {
	return l.focused == l.root.next
}

func (l *ItemList) Focused() Item {
	if l.focused != nil {
		return l.focused.Data
	}

	return nil
}

func (l *ItemList) Len() int {
	return l.count
}

func (l *ItemList) All() []Item {
	var members []Item

	for n := l.root.next; n != &l.root; n = n.next {
		members = append(members, n.Data)
	}

	return members
}

func (l *ItemList) String() string {
	if l.root.next == &l.root {
		return "<nil>"
	} else {
		var s []string

		for n := l.root.next; n != &l.root; n = n.next {
			if l.focused == n {
				s = append(s, fmt.Sprintf("<%s>", n.Data))
			} else {
				s = append(s, fmt.Sprintf("-%s-", n.Data))
			}
		}

		return strings.Join(s, " ")
	}
}

func NewItemList() *ItemList {
	l := new(ItemList)

	l.root.list = l
	l.root.prev = &l.root
	l.root.next = &l.root

	return l
}

func (l *ItemList) link(node, after *ItemNode) *ItemNode {
	next := after.next
	after.next = node
	node.prev = after
	node.next = next
	next.prev = node
	node.list = l
	l.count += 1

	return node
}

func (l *ItemList) insert(node, after *ItemNode) *ItemNode {
	if l.focused == nil {
		l.focused = node
	}

	return l.link(node, after)
}

func (l *ItemList) insertData(data Item, after *ItemNode) *ItemNode {
	return l.insert(&ItemNode{Data: data}, after)
}

func (l *ItemList) unlink(node *ItemNode) *ItemNode {
	node.prev.next = node.next
	node.next.prev = node.prev
	node.prev = nil
	node.next = nil
	node.list = nil
	l.count -= 1

	return node
}

func (l *ItemList) remove(node *ItemNode) *ItemNode {
	if node == l.focused {
		if l.focused.prev == &l.root {
			l.focused = l.focused.next
		} else {
			l.focused = l.focused.prev
		}
	}

	return l.unlink(node)
}

func (l *ItemList) swap(node, with *ItemNode) {
	if l.focused == node {
		l.focused = with
	}

	data := node.Data
	node.Data = with.Data
	with.Data = data
}

//
// INSERT
//

func (l *ItemList) Insert(data Item) *ItemNode {
	return l.insertData(data, l.root.prev)
}

func (l *ItemList) InsertAfterFocus(data Item) *ItemNode {
	return l.insertData(data, l.focused)
}

//
// NODE
//

func (l *ItemList) NodeFunc(test func(Item) bool, operation func(*ItemNode)) {
	for n := l.root.next; n != &l.root; n = n.next {
		if test(n.Data) {
			operation(n)
			return
		}
	}
}

func (l *ItemList) NodeMatch(target Item, operation func(*ItemNode)) {
	l.NodeFunc(func(data Item) bool { return data == target }, operation)
}

//
// SELECT
//

func (l *ItemList) Select(count int) Item {
	if l.focused == nil || count == 0 {
		return nil
	}

	node := l.focused
	for i := 0; i < int(math.Abs(float64(count))); i++ {
		if count < 0 {
			node = node.prev

			if node == &l.root {
				node = node.prev
			}

			if node == l.focused {
				node = node.prev
			}
		} else {
			node = node.next

			if node == &l.root {
				node = node.next
			}

			if node == l.focused {
				node = node.next
			}
		}
	}

	return node.Data
}

//
// FOCUS
//

func (l *ItemList) FocusNext() {
	if l.focused != nil {
		if l.focused.next == &l.root {
			l.focused = l.root.next
		} else {
			l.focused = l.focused.next
		}
	}
}

func (l *ItemList) FocusPrev() {
	if l.focused != nil {
		if l.focused.prev == &l.root {
			l.focused = l.root.prev
		} else {
			l.focused = l.focused.prev
		}
	}
}

func (l *ItemList) Focus(count int) {
	if l.focused == nil || count == 0 {
		return
	}

	for i := 0; i < int(math.Abs(float64(count))); i++ {
		if count < 0 {
			l.FocusPrev()
		} else {
			l.FocusNext()
		}
	}
}

func (l *ItemList) FocusFunc(test func(Item) bool) {
	l.NodeFunc(test, func(node *ItemNode) {
		l.focused = node
	})
}

func (l *ItemList) FocusMatch(target Item) {
	l.FocusFunc(func(data Item) bool {
		return data == target
	})
}

//
// REMOVE
//

func (l *ItemList) RemoveFocus() {
	if l.focused != nil {
		l.remove(l.focused)
	}
}

func (l *ItemList) RemoveFunc(test func(Item) bool) {
	l.NodeFunc(test, func(node *ItemNode) {
		l.remove(node)
	})
}

func (l *ItemList) RemoveMatch(target Item) {
	l.RemoveFunc(func(data Item) bool {
		return data == target
	})
}

//
// SWAP
//

func (l *ItemList) MoveFocusNext() {
	if l.focused != nil {
		if l.focused.next == &l.root {
			l.swap(l.focused, l.root.next)
		} else {
			l.swap(l.focused, l.focused.next)
		}
	}
}

func (l *ItemList) MoveFocusPrev() {
	if l.focused != nil {
		if l.focused.prev == &l.root {
			l.swap(l.focused, l.root.prev)
		} else {
			l.swap(l.focused, l.focused.prev)
		}
	}
}

func (l *ItemList) MoveFocusFront() {
	if l.focused != nil {
		l.swap(l.focused, l.root.next)
	}
}

func (l *ItemList) MoveFocus(count int) {
	if l.focused == nil || count == 0 {
		return
	}

	for i := 0; i < int(math.Abs(float64(count))); i++ {
		if count < 0 {
			l.MoveFocusPrev()
		} else {
			l.MoveFocusNext()
		}
	}
}

func (l *ItemList) MoveFocusFunc(test func(Item) bool) {
	l.NodeFunc(test, func(node *ItemNode) {
		//focused := l.focused
		l.swap(l.focused, node)
		//l.focused = focused
	})
}

func (l *ItemList) MoveFocusMatch(target Item) {
	l.MoveFocusFunc(func(data Item) bool {
		return data == target
	})
}

// FIXME vurdere
func (l *ItemList) SwapFront(node *ItemNode) {
	l.swap(node, l.root.next)
}
