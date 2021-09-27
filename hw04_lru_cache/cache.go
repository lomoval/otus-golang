package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mutex    sync.Mutex
}

type cacheItem struct {
	key   string
	value interface{}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (l *lruCache) Set(key Key, value interface{}) bool {
	if i, ok := l.items[key]; ok {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		i.Value = value
		l.queue.MoveToFront(i)
		l.items[key].Value = cacheItem{key: string(key), value: value}
		return true
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.removeObsolete()
	l.items[key] = l.queue.PushFront(cacheItem{key: string(key), value: value})
	return false
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if i, ok := l.items[key]; ok {
		l.queue.MoveToFront(i)
		return i.Value.(cacheItem).value, true
	}
	return nil, false
}

func (l *lruCache) Clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.queue = NewList()
	l.items = make(map[Key]*ListItem, l.capacity)
}

func (l *lruCache) full() bool {
	return len(l.items) >= l.capacity
}

func (l *lruCache) removeObsolete() {
	if l.full() {
		e := l.queue.Back()
		delete(l.items, Key(e.Value.(cacheItem).key))
		l.queue.Remove(e)
	}
}
