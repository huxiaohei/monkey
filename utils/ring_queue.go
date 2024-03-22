package utils

type RingQueue[V any] struct {
	capacity uint64
	size     uint64
	head     uint64
	tail     uint64
	data     []V
}

func NewRingQueue[V any](capacity uint64) *RingQueue[V] {
	return &RingQueue[V]{
		capacity: capacity,
		size:     0,
		head:     0,
		tail:     0,
		data:     make([]V, capacity),
	}
}

func (rq *RingQueue[V]) Push(v V, drop bool) bool {
	if rq.size == rq.capacity {
		if drop {
			rq.head = (rq.head + 1) % rq.capacity
			rq.size--
		} else {
			return false
		}
	}
	rq.data[rq.tail] = v
	rq.tail = (rq.tail + 1) % rq.capacity
	rq.size++
	return true
}

func (rq *RingQueue[V]) Pop() (*V, bool) {
	if rq.size == 0 {
		return nil, false
	}
	v := rq.data[rq.head]
	rq.head = (rq.head + 1) % rq.capacity
	rq.size--
	return &v, true
}

func (rq *RingQueue[V]) Size() uint64 {
	return rq.size
}

func (rq *RingQueue[V]) Capacity() uint64 {
	return rq.capacity
}
