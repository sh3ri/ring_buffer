package ring_buffer_test

import (
	"circular_buffer/ring_buffer"
	"context"
	"math/rand"
	"testing"
	"time"
)

type TestStruct struct{ value int }

func TestRingBuffer_WriteTimeout(t *testing.T) {
	size := 4
	rb := ring_buffer.NewRingBuffer[int](size)
	for i := 0; i < size; i++ {
		err := rb.Write(1, time.Nanosecond)
		if err != nil {
			t.Fatalf("could not write to ring in 1 microsecond")
		}
	}
	err := rb.Write(1, 20*time.Millisecond)
	if err == nil {
		t.Fatalf("write should not have succeeded")
	}
	if err != context.DeadlineExceeded {
		t.Fatalf("wrong error type %T", err)
	}
}

func writeToBuffer(t *testing.T, rb *ring_buffer.RingBuffer[*TestStruct], inputList []*TestStruct) *TestStruct {
	element := &TestStruct{value: rand.Int()}
	err := rb.Write(element, time.Nanosecond)
	if err != nil {
		t.Fatalf("write returned %v", err)
	}
	return element
}

func readFromBuffer(t *testing.T, rb *ring_buffer.RingBuffer[*TestStruct], outputList []*TestStruct) *TestStruct {
	element, err := rb.Read()
	if err != nil {
		t.Fatalf("read returned %v", err)
	}
	return element
}

func TestRingBuffer_ReadOrder(t *testing.T) {
	size := 4
	rb := ring_buffer.NewRingBuffer[*TestStruct](size)
	inputList := make([]*TestStruct, 0)
	outputList := make([]*TestStruct, 0)
	for i := 0; i < size; i++ {
		inputList = append(inputList, writeToBuffer(t, rb, inputList))
	}
	for i := 0; i < size-1; i++ {
		outputList = append(outputList, readFromBuffer(t, rb, outputList))
	}
	for i := 0; i < size-1; i++ {
		inputList = append(inputList, writeToBuffer(t, rb, inputList))
	}
	for i := 0; i < size; i++ {
		outputList = append(outputList, readFromBuffer(t, rb, outputList))
	}
	for i, outputElement := range outputList {
		if outputElement != inputList[i] {
			t.Fatalf("value read from the buffer isn't equal to the value written to it. expected %v, got %v", inputList[i], outputElement)
		}
	}
}
