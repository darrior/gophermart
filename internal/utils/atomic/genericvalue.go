// Package atomic provides some useful atomic.Value wrappers
package atomic

import (
	"sync/atomic"
)

type GenericValue[T any] struct {
	v atomic.Value
}

func NewGenericValue[T any](val T) GenericValue[T] {
	t := GenericValue[T]{
		v: atomic.Value{},
	}

	t.v.Store(val)

	return t
}

func (t *GenericValue[T]) Load() T {
	return t.v.Load().(T)
}

func (t *GenericValue[T]) Store(v T) {
	t.v.Store(v)
}
