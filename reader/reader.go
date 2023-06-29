package reader

import (
	"circular_buffer/models"
	"circular_buffer/ring_buffer"
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"
)

type Reader[T models.ANY] struct {
	buffer              *ring_buffer.RingBuffer[T]
	doneChan            chan struct{}
	stopIfBufferIsEmpty bool
	wg                  sync.WaitGroup
}

func NewReader[T models.ANY](buffer *ring_buffer.RingBuffer[T]) *Reader[T] {
	return &Reader[T]{buffer: buffer, doneChan: make(chan struct{}), stopIfBufferIsEmpty: false, wg: sync.WaitGroup{}}
}

func (r *Reader[T]) StopIfBufferIsEmpty() {
	r.stopIfBufferIsEmpty = true
}

func (r *Reader[T]) Done() chan struct{} {
	return r.doneChan
}

func (r *Reader[T]) processData(data *T) {
	marshalled, err := sonic.Marshal(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", string(marshalled))
	r.wg.Done()
}

func (r *Reader[T]) Start() {
	for {
		data, err := r.buffer.Read()
		if err == ring_buffer.ErrClosed {
			r.wg.Wait()
			close(r.doneChan)
		}
		if err == ring_buffer.ErrIsEmpty {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		r.wg.Add(1)
		go r.processData(data)
	}
}
