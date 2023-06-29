package ring_buffer

import (
	"circular_buffer/models"
	"circular_buffer/util"
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

var ErrIsEmpty = errors.New("can't read from empty buffer")
var ErrIsFull = errors.New("buffer is full")
var ErrClosed = errors.New("buffer is closed")
var ErrNotEnoughElements = errors.New("there aren't enough elements in the buffer")

type RingBuffer[T models.ANY] struct {
	sync.Mutex
	buffer        []*T
	read_pointer  int
	write_pointer int
	isFull        bool
	size          int
	done          bool
	semaphore     *semaphore.Weighted
}

func NewRingBuffer[T models.ANY](size int) *RingBuffer[T] {
	return &RingBuffer[T]{
		buffer:        make([]*T, size),
		read_pointer:  0,
		write_pointer: 0,
		isFull:        false,
		size:          size,
		done:          false,
		semaphore:     semaphore.NewWeighted(int64(size)),
	}
}

func (rb *RingBuffer[T]) Close() {
	rb.Lock()
	defer rb.Unlock()
	rb.done = true
}

func (rb *RingBuffer[T]) ConfigureSize(newSize int) error {
	rb.Lock()
	defer rb.Unlock()

	if newSize < rb.occupiedSlotsCount() {
		return errors.New("new size is is less than the number of elements currently in the buffer")
	}
	if rb.size == newSize {
		return nil
	}

	occupiedSlotsCount := rb.occupiedSlotsCount()
	newBuffer := make([]*T, newSize)
	if rb.isFull {
		copy(newBuffer[:occupiedSlotsCount], rb.buffer)
	} else if rb.read_pointer <= rb.write_pointer {
		copy(newBuffer[:occupiedSlotsCount], rb.buffer[rb.read_pointer:rb.write_pointer])
	} else {
		firstSectionSize := rb.size - rb.read_pointer
		secondSectionSize := occupiedSlotsCount - firstSectionSize
		copy(newBuffer[:firstSectionSize], rb.buffer[rb.read_pointer:])
		copy(newBuffer[firstSectionSize:firstSectionSize+secondSectionSize], rb.buffer[:rb.write_pointer])
	}

	rb.size = newSize
	rb.read_pointer = 0
	rb.write_pointer = occupiedSlotsCount % newSize
	rb.isFull = occupiedSlotsCount == newSize
	rb.buffer = newBuffer
	rb.semaphore = semaphore.NewWeighted(int64(newSize))
	return nil
}

func (rb *RingBuffer[T]) nextIndex(currentIndex, steps int) int {
	return (currentIndex + steps) % rb.size
}

func (rb *RingBuffer[T]) occupiedSlotsCount() int {
	if rb.isFull {
		return rb.size
	}
	if rb.write_pointer >= rb.read_pointer {
		return rb.write_pointer - rb.read_pointer
	}
	return rb.size - (rb.read_pointer - rb.write_pointer - 1)
}

func (rb *RingBuffer[T]) freeSlotsCount() int {
	return rb.size - rb.occupiedSlotsCount()
}

func (rb *RingBuffer[T]) IsEmpty() bool {
	return rb.read_pointer == rb.write_pointer && !rb.isFull
}

func (rb *RingBuffer[T]) Read() (*T, error) {
	res, err := rb.ReadMany(1)
	if err != nil {
		return new(T), err
	}
	return res[0], nil
}

func (rb *RingBuffer[T]) ReadMany(count int) ([]*T, error) {
	if count == 0 {
		return []*T{}, nil
	}

	if rb.IsEmpty() {
		if rb.done {
			return nil, ErrClosed
		}
		return nil, ErrIsEmpty
	}

	rb.Lock()
	defer rb.Unlock()

	occupiedSlotsCount := rb.occupiedSlotsCount()
	resultSize := util.Min(count, occupiedSlotsCount)
	result := make([]*T, resultSize)
	if rb.read_pointer+resultSize <= rb.size {
		copy(result, rb.buffer[rb.read_pointer:rb.read_pointer+resultSize])
	} else {
		firstSectionSize := rb.size - rb.read_pointer
		secondSectionSize := resultSize - firstSectionSize
		copy(result[:firstSectionSize], rb.buffer[rb.read_pointer:])
		copy(result[firstSectionSize:], rb.buffer[:secondSectionSize])
	}

	rb.read_pointer = rb.nextIndex(rb.read_pointer, resultSize)
	rb.isFull = false
	rb.semaphore.Release(int64(resultSize))

	if resultSize < count {
		return result, ErrNotEnoughElements
	}
	return result, nil
}

func (rb *RingBuffer[T]) Write(element *T, timeout time.Duration) error {
	if rb.done {
		return ErrClosed
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	rb.semaphore.Acquire(ctx, 1)
	rb.Lock()
	defer rb.Unlock()

	rb.buffer[rb.write_pointer] = element
	rb.write_pointer = rb.nextIndex(rb.write_pointer, 1)
	rb.isFull = rb.write_pointer == rb.read_pointer
	return nil
}
