package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

func (li *ListItem) resetLinks() {
	li.Next = nil
	li.Prev = nil
}

type list struct {
	len       int
	firstNode *ListItem
	lastNode  *ListItem
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.firstNode
}

func (l *list) Back() *ListItem {
	return l.lastNode
}

func (l *list) PushFront(v interface{}) *ListItem {
	newItem := ListItem{Value: v}
	if l.empty() {
		l.lastNode = &newItem
		l.firstNode = &newItem
		l.len = 1
		return l.firstNode
	}

	newItem.Next = l.firstNode
	l.firstNode.Prev = &newItem
	l.firstNode = &newItem

	l.len++
	return l.firstNode
}

func (l *list) PushBack(v interface{}) *ListItem {
	item := ListItem{Value: v}
	if l.empty() {
		return l.PushFront(v)
	}

	l.lastNode.Next = &item
	item.Prev = l.lastNode
	l.lastNode = &item

	l.len++
	return l.lastNode
}

func (l *list) Remove(i *ListItem) {
	switch {
	case l.firstNode == i && l.lastNode == i:
		l.firstNode = nil
		l.lastNode = nil
	case l.firstNode == i:
		l.firstNode = i.Next
	case l.lastNode == i:
		l.lastNode = i.Prev
		l.lastNode.Next = nil
	default:
		i.Next.Prev = i.Prev
		i.Prev.Next = i.Next
	}
	l.len--
	i.resetLinks()
}

func (l *list) MoveToFront(i *ListItem) {
	switch {
	case l.firstNode == i:
		return
	case l.lastNode == i:
		l.lastNode = i.Prev
		l.lastNode.Next = nil
	default:
		i.Next.Prev = i.Prev
		i.Prev.Next = i.Next
	}
	l.firstNode.Prev = i
	i.Next = l.firstNode
	i.Prev = nil
	l.firstNode = i
}

func (l *list) empty() bool {
	return l.len == 0
}

func NewList() List {
	return new(list)
}
