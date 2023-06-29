package main

import (
	"bufio"
	"circular_buffer/models"
	"circular_buffer/reader"
	"circular_buffer/ring_buffer"
	"fmt"
	"os"
	"strings"

	"github.com/bytedance/sonic"
)

func WriteFileToBuffer[T models.ANY](path string, buffer *ring_buffer.RingBuffer[T]) {
	defer buffer.Close()
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rawData := scanner.Text()
		if strings.Trim(rawData, " ") == "" {
			break
		}
		var data T
		err = sonic.UnmarshalString(rawData, &data)
		if err != nil {
			panic(err)
		}
		err = buffer.Write(&data)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
	}
}

func main() {
	buffer := ring_buffer.NewRingBuffer[*models.Data](500)
	reader := reader.NewReader(buffer)
	go reader.Start()
	WriteFileToBuffer[*models.Data]("data.json", buffer)
	<-reader.Done()
	fmt.Println("done")
}
