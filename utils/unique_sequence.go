package utils

import (
	"sync/atomic"
)

type SessionId uint64

type UniqueSequence struct {
	nextId uint64
}

func NewUniqueSequence(minId uint64) *UniqueSequence {
	return &UniqueSequence{nextId: minId}
}

func (us *UniqueSequence) GetNewId() uint64 {
	id := atomic.AddUint64(&us.nextId, 1)
	return id
}

func (us *UniqueSequence) GetNewSequence(step uint64) uint64 {
	id := atomic.AddUint64(&us.nextId, step)
	return id
}

func (us *UniqueSequence) GetNewSessionId() SessionId {
	id := atomic.AddUint64(&us.nextId, 1)
	return SessionId(id)
}
