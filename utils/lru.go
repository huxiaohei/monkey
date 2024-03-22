package utils

import "sync"

type LRUNode[K comparable, V any] struct {
	key  K
	data *V
	prev *LRUNode[K, V]
	next *LRUNode[K, V]
}

type LRULinkedList[K comparable, V any] struct {
	head *LRUNode[K, V]
	tail *LRUNode[K, V]
}

func (list *LRULinkedList[K, V]) append(node *LRUNode[K, V]) {
	if list.head == nil {
		list.head = node
		list.tail = node
	} else {
		node.next = list.head
		list.head.prev = node
		list.head = node
	}
}

func (list *LRULinkedList[K, V]) remove(node *LRUNode[K, V]) {
	if node == list.head {
		list.head = node.next
	} else if node == list.tail {
		list.tail = node.prev
	} else {
		node.prev.next = node.next
		node.next.prev = node.prev
	}
}

type LRU[K comparable, V any] struct {
	capacity   uint64
	size       uint64
	nodeMap    map[K]*LRUNode[K, V]
	linkedList LRULinkedList[K, V]
	lock       sync.RWMutex
}

func NewLRU[K comparable, V any](capacity uint64) *LRU[K, V] {
	return &LRU[K, V]{
		capacity:   capacity,
		size:       0,
		nodeMap:    make(map[K]*LRUNode[K, V]),
		linkedList: LRULinkedList[K, V]{},
		lock:       sync.RWMutex{},
	}
}

func (lru *LRU[K, V]) Contain(key K) bool {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	_, ok := lru.nodeMap[key]
	return ok
}

func (lru *LRU[K, V]) Get(key K) (*V, bool) {
	lru.lock.RLock()
	defer lru.lock.RUnlock()

	if node, ok := lru.nodeMap[key]; ok {
		return node.data, ok
	}
	return nil, false
}

func (lru *LRU[K, V]) Put(key K, value V) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if node, ok := lru.nodeMap[key]; ok {
		node.data = &value
		lru.linkedList.remove(node)
		lru.linkedList.append(node)
	} else {
		if lru.size >= lru.capacity {
			delete(lru.nodeMap, lru.linkedList.head.key)
			lru.linkedList.remove(lru.linkedList.head)
			lru.size--
		}
		node := &LRUNode[K, V]{key: key, data: &value}
		lru.nodeMap[key] = node
		lru.linkedList.append(node)
		lru.size++
	}
}

func (lru *LRU[K, V]) Remove(key K) {
	lru.lock.Lock()
	defer lru.lock.Unlock()

	if node, ok := lru.nodeMap[key]; ok {
		delete(lru.nodeMap, key)
		lru.linkedList.remove(node)
		lru.size--
	}
}

func (lru *LRU[K, V]) Len() uint64 {
	lru.lock.RLock()
	defer lru.lock.RUnlock()

	return lru.size
}

func (lru *LRU[K, V]) Cap() uint64 {
	lru.lock.RLock()
	defer lru.lock.RUnlock()

	return lru.capacity
}
