package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}

func TestListRemoveAllFromEnd(t *testing.T) {
	l := NewList()
	expected := []int{1, 2, 3, 4, 5}

	for i := len(expected) - 1; i >= 0; i-- {
		l.PushFront(expected[i])
	}
	require.Equal(t, len(expected), l.Len())

	for i := 1; i < len(expected); i++ {
		l.Remove(l.Back())
		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, expected[:len(expected)-i], elems)
	}
}

func TestListRemoveAllFromStart(t *testing.T) {
	l := NewList()
	expected := []int{1, 2, 3, 4, 5}

	for i := len(expected) - 1; i >= 0; i-- {
		l.PushFront(expected[i])
	}
	require.Equal(t, len(expected), l.Len())

	for i := 1; i < len(expected); i++ {
		l.Remove(l.Front())
		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, expected[i:], elems)
	}
}

func TestListAddOneToFront(t *testing.T) {
	l := NewList()

	e := l.PushFront(10)
	require.Equal(t, 1, l.Len())
	require.Equal(t, 10, e.Value)
	require.Equal(t, e, l.Front())
	require.Equal(t, e, l.Back())
}

func TestListAddOneToBack(t *testing.T) {
	l := NewList()

	e := l.PushBack(10)
	require.Equal(t, 1, l.Len())
	require.Equal(t, 10, e.Value)
	require.Equal(t, e, l.Front())
	require.Equal(t, e, l.Back())
}

func TestListMoveOneToFront(t *testing.T) {
	l := NewList()

	e := l.PushFront(10)
	require.Equal(t, 1, l.Len())
	require.Equal(t, 10, e.Value)
	require.Equal(t, e, l.Front())
	require.Equal(t, e, l.Back())

	l.MoveToFront(e)
	require.Equal(t, 1, l.Len())
	require.Equal(t, 10, e.Value)
	require.Equal(t, e, l.Front())
	require.Equal(t, e, l.Back())
}

func TestListRemoveOne(t *testing.T) {
	l := NewList()

	e := l.PushFront(10)
	require.Equal(t, 1, l.Len())
	require.Equal(t, 10, e.Value)

	l.Remove(e)

	require.Equal(t, 0, l.Len())
	require.Equal(t, 10, e.Value)
	require.Nil(t, l.Front())
	require.Nil(t, l.Back())
}

func TestListRemoveFromMiddle(t *testing.T) {
	l := NewList()

	last := l.PushFront(10)
	middle := l.PushFront(20)
	first := l.PushFront(30)
	require.Equal(t, 3, l.Len())
	require.Equal(t, 30, first.Value)
	require.Equal(t, 20, middle.Value)
	require.Equal(t, 10, last.Value)

	l.Remove(middle)

	require.Equal(t, 2, l.Len())
	require.Equal(t, 30, first.Value)
	require.Equal(t, 20, middle.Value)
	require.Equal(t, 10, last.Value)
	require.Equal(t, first, l.Front())
	require.Equal(t, last, l.Back())

	elems := make([]int, 0, l.Len())
	for i := l.Front(); i != nil; i = i.Next {
		elems = append(elems, i.Value.(int))
	}
	require.Equal(t, []int{30, 10}, elems)

	elems = make([]int, 0, l.Len())
	for i := l.Back(); i != nil; i = i.Prev {
		elems = append(elems, i.Value.(int))
	}
	require.Equal(t, []int{10, 30}, elems)
}
