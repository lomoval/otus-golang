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
}

func TestCacheRemoving(t *testing.T) {
	t.Run("RemoveObsolete", func(t *testing.T) {
		c := NewCache(3)
		c.Set("1", 100)
		c.Set("2", 200)
		c.Set("3", 300)
		c.Set("4", 400)

		val, ok := c.Get("1")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("2")
		require.True(t, ok)
		require.Equal(t, 200, val)
		val, ok = c.Get("3")
		require.True(t, ok)
		require.Equal(t, 300, val)
		val, ok = c.Get("4")
		require.True(t, ok)
		require.Equal(t, 400, val)
	})

	t.Run("RemoveObsoleteAfterSetting", func(t *testing.T) {
		c := NewCache(3)
		c.Set("1", 100)
		c.Set("2", 200)
		c.Set("3", 300)

		c.Set("1", 100)
		c.Set("3", 300)

		c.Set("4", 400)

		val, ok := c.Get("2")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("1")
		require.True(t, ok)
		require.Equal(t, 100, val)
		val, ok = c.Get("3")
		require.True(t, ok)
		require.Equal(t, 300, val)
		val, ok = c.Get("4")
		require.True(t, ok)
		require.Equal(t, 400, val)
	})

	t.Run("RemoveObsoleteAfterGetting", func(t *testing.T) {
		c := NewCache(3)
		c.Set("1", 100)
		c.Set("2", 200)
		c.Set("3", 300)

		c.Get("1")
		c.Get("3")

		c.Set("4", 400)

		val, ok := c.Get("2")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("1")
		require.True(t, ok)
		require.Equal(t, 100, val)
		val, ok = c.Get("3")
		require.True(t, ok)
		require.Equal(t, 300, val)
		val, ok = c.Get("4")
		require.True(t, ok)
		require.Equal(t, 400, val)
	})

	t.Run("RemoveSeveralObsolete", func(t *testing.T) {
		c := NewCache(3)
		c.Set("1", 100)
		c.Set("2", 200)
		c.Set("3", 300)

		c.Set("1", "100")
		c.Get("3")

		c.Set("4", 400)
		c.Set("5", 500)

		val, ok := c.Get("2")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("1")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("3")
		require.True(t, ok)
		require.Equal(t, 300, val)
		val, ok = c.Get("4")
		require.True(t, ok)
		require.Equal(t, 400, val)
		val, ok = c.Get("5")
		require.True(t, ok)
		require.Equal(t, 500, val)
	})
}

func TestCacheClean(t *testing.T) {
	c := NewCache(3)
	c.Set("1", 100)
	c.Set("2", 200)
	c.Set("3", 300)

	val, ok := c.Get("1")
	require.True(t, ok)
	require.Equal(t, 100, val)
	val, ok = c.Get("2")
	require.True(t, ok)
	require.Equal(t, 200, val)
	val, ok = c.Get("3")
	require.True(t, ok)
	require.Equal(t, 300, val)

	c.Clear()

	val, ok = c.Get("1")
	require.False(t, ok)
	require.Nil(t, val)
	val, ok = c.Get("2")
	require.False(t, ok)
	require.Nil(t, val)
	val, ok = c.Get("3")
	require.False(t, ok)
	require.Nil(t, val)
}

func TestCacheMultithreading(t *testing.T) {
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
