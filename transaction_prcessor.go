package main

import (
	"circular_buffer/models"
	"circular_buffer/reader"
	"circular_buffer/ring_buffer"
)

func main() {
	buffer := ring_buffer.NewRingBuffer[*models.Data](4)
	reader := reader.NewReader(buffer)
	go buffer.IngestFile("data.json")
	go reader.Start()
	<-reader.Done()
}
