// Package syncqueue provides threadsafe queue.
package syncqueue

import "sync"

type SyncQueue[T any] struct {
	lock  *sync.Mutex
	queue []T
}

func NewSyncQueue[T any]() *SyncQueue[T] {
	return &SyncQueue[T]{
		lock:  &sync.Mutex{},
		queue: make([]T, 0),
	}
}

func (s *SyncQueue[T]) Pop() (T, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.queue) == 0 {
		var ret T
		return ret, false
	}

	first := s.queue[0]
	s.queue = s.queue[1:]
	return first, true
}

func (s *SyncQueue[T]) Push(element T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.queue = append(s.queue, element)
}

func (s *SyncQueue[t]) IsEmpty() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return len(s.queue) == 0
}
