// Package syncqueue provides threadsafe queue.
package syncqueue

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncQueue_Pop(t *testing.T) {
	tests := []struct {
		name  string
		args  []int
		want  int
		want1 bool
	}{
		{
			name:  "Empty Queue",
			args:  []int{},
			want:  0,
			want1: false,
		},
		{
			name:  "One element",
			args:  []int{1},
			want:  1,
			want1: true,
		},
		{
			name:  "Some elements",
			args:  []int{1, 2, 3},
			want:  1,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SyncQueue[int]{
				lock:  &sync.Mutex{},
				queue: tt.args,
			}

			got, got1 := s.Pop()

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func TestSyncQueue_Push(t *testing.T) {
	type args struct {
		element int
	}
	tests := []struct {
		name  string
		queue []int
		args  args
		want  []int
	}{
		{
			name:  "Empty queue",
			queue: []int{},
			args: args{
				element: 10,
			},
			want: []int{10},
		},
		{
			name:  "Non-empty queue",
			queue: []int{1},
			args: args{
				element: 10,
			},
			want: []int{1, 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SyncQueue[int]{
				lock:  &sync.Mutex{},
				queue: tt.queue,
			}

			s.Push(tt.args.element)

			assert.Equal(t, tt.want, s.queue)
		})
	}
}

func TestSyncQueue_IsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		queue []int
		want  bool
	}{
		{
			name:  "Empty queue",
			queue: []int{},
			want:  true,
		},
		{
			name:  "Non-empty queue",
			queue: []int{1},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SyncQueue[int]{
				lock:  &sync.Mutex{},
				queue: tt.queue,
			}

			assert.Equal(t, tt.want, s.IsEmpty())
		})
	}
}
