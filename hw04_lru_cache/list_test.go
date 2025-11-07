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

func TestList2(t *testing.T) {
	t.Run("moving", func(t *testing.T) {
		l := NewList()

		l.PushBack(16) // [16]
		l.PushBack(25) // [16, 25]
		l.PushFront(9) // [9, 16, 25]
		l.PushFront(4) // [4, 9, 16, 25]
		l.PushFront(1) // [1, 4, 9, 16, 25]
		l.PushBack(36) // [1, 4, 9, 16, 25, 36]
		l.PushBack(49) // [1, 4, 9, 16, 25, 36, 49]
		require.Equal(t, 7, l.Len())

		middle := l.Front().Next.Next.Next // 16
		l.Remove(middle)                   // [1, 4, 9, 25, 36, 49]
		require.Equal(t, 6, l.Len())

		l.MoveToFront(l.Front().Next.Next.Next.Next.Next) // [49, 1, 4, 9, 25, 36]
		l.MoveToFront(l.Front().Next.Next.Next.Next.Next) // [36, 49, 1, 4, 9, 25]
		l.MoveToFront(l.Front().Next.Next.Next.Next.Next) // [25, 36, 49, 1, 4, 9]
		l.MoveToFront(l.Front().Next.Next.Next)           // [1, 25, 36, 49, 4, 9]
		require.Equal(t, 6, l.Len())
		require.Equal(t, 1, l.Front().Value)
		require.Equal(t, 25, l.Front().Next.Value)
		require.Equal(t, 36, l.Front().Next.Next.Value)
		require.Equal(t, 49, l.Front().Next.Next.Next.Value)
		require.Equal(t, 4, l.Front().Next.Next.Next.Next.Value)
		require.Equal(t, 9, l.Back().Value)

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{1, 25, 36, 49, 4, 9}, elems)
	})
}
