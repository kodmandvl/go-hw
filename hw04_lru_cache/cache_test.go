package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		// Проверим на выталкивание из кэша, размер будет 3:
		c := NewCache(3)

		// Наполняем кэш:
		wasInCache := c.Set("page0block0", "something")
		require.False(t, wasInCache)
		wasInCache = c.Set("page3block7", "aaa")
		require.False(t, wasInCache)
		wasInCache = c.Set("page5block4", "bbb")
		require.False(t, wasInCache)

		// Проверим на выталкивание по превышению размера:
		wasInCache = c.Set("page6block2", "ссс")
		require.False(t, wasInCache)
		// Теперь элемент page0block0 должен отсутствовать:
		v, ok := c.Get("page0block0")
		require.False(t, ok)
		require.Nil(t, v)

		// Проверим на выталкивание наиболее давно используемого элемента:
		wasInCache = c.Set("page6block2", "new_val_c")
		require.True(t, wasInCache)
		wasInCache = c.Set("page5block4", "new_val_b")
		require.True(t, wasInCache)
		wasInCache = c.Set("page3block7", "new_val_a")
		require.True(t, wasInCache)
		v, ok = c.Get("page6block2")
		require.True(t, ok)
		require.Equal(t, "new_val_c", v)
		wasInCache = c.Set("page7block7", "ggg")
		require.False(t, wasInCache)
		// Теперь элемент page5block4 должен отсутствовать:
		v, ok = c.Get("page5block4")
		require.False(t, ok)
		require.Nil(t, v)
	})
}

func TestCacheMultithreading(_ *testing.T) {
	// t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
