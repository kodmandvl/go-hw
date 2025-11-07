package hw04lrucache

import (
	"sync"
)

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	lock     sync.Mutex
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		lock:     sync.Mutex{},
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

type cacheElem struct {
	k Key
	v interface{}
}

// Чтобы сделать кэш горутино-безопасным, добавим sync.Mutex в структуру кэша
// и вызов Lock в начале функций, а Unlock - в конце, через defer.
// Это позволит защитить кэш, т.к. Mutex занять может только одна горутина.
// В таком случае наш кэш будет уже каким-то подобием in-memory БД с защищенным доступом

// Set(key Key, value interface{}) bool // Добавить значение в кэш по ключу.
func (c *lruCache) Set(key Key, value interface{}) bool {
	// В начале ставим блокировку (mutex):
	c.lock.Lock()
	// В конце освобождаем блокировку (mutex):
	defer c.lock.Unlock()
	// Пытаемся получить элемент по ключу:
	elem, ok := c.items[key]
	if ok {
		// Если нашли ключ, то переприсваиваем значение и перемещаем наверх:
		elem.Value = &cacheElem{
			k: key,
			v: value,
		}
		c.queue.MoveToFront(elem)
		return true // ключ был, обновили имеющийся элемнт в кэше
	}
	// Если не нашли ключ, то добавляем новый элемент в кэш:
	newElem := c.queue.PushFront(&cacheElem{
		k: key,
		v: value,
	})
	c.items[key] = newElem
	// При этом если превысили ёмкость кэша, то удаляем последний элемент:
	if c.queue.Len() > c.capacity {
		lastElem := c.queue.Back()
		c.queue.Remove(lastElem)                       // удаляем из списка
		delete(c.items, lastElem.Value.(*cacheElem).k) // удаляем из мапы
	}
	return false // ключа не было, добавили новый элемент в кэш
}

// Get(key Key) (interface{}, bool) // Получить значение из кэша по ключу.
func (c *lruCache) Get(key Key) (interface{}, bool) {
	// В начале ставим блокировку (mutex):
	c.lock.Lock()
	// В конце освобождаем блокировку (mutex):
	defer c.lock.Unlock()
	// Пытаемся получить элемент по ключу:
	elem, ok := c.items[key]
	if ok {
		// Если нашли ключ, то перемещаем его наверх, а также возвращаем значение и true:
		c.queue.MoveToFront(elem)
		return elem.Value.(*cacheElem).v, true
	}
	// Если не нашли ключ, то возвращаем nil и false:
	return nil, false
}

// Clear() // Очистить кэш.
func (c *lruCache) Clear() {
	// В начале ставим блокировку (mutex):
	c.lock.Lock()
	// В конце освобождаем блокировку (mutex):
	defer c.lock.Unlock()
	// Очищаем кэш (приводим к состоянию нового пустого кэша с текущей capacity):
	c.items = make(map[Key]*ListItem, c.capacity) // новая чистая мапа, но с текущей capacity (см. func NewCache)
	c.queue = NewList()                           // новый пустой список (см. func NewCache)
}
