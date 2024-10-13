package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// Константы для настройки размера буфера и интервала времени
const (
	bufferSize    = 5
	flushInterval = 5 * time.Second
)

// кольцевой буфер
type CircularBuffer struct {
	data  []int
	size  int
	head  int
	tail  int
	count int
	mutex sync.Mutex
}

// создание нового буфера
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		data:  make([]int, size),
		size:  size,
		head:  0,
		tail:  0,
		count: 0,
	}
}

// добавление элемента в буфер
func (cb *CircularBuffer) Push(value int) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.data[cb.tail] = value
	cb.tail = (cb.tail + 1) % cb.size
	if cb.count == cb.size {
		cb.head = (cb.head + 1) % cb.size
	} else {
		cb.count++
	}
}

// выводим содержимое буфера и чистим его
func (cb *CircularBuffer) Flush() []int {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.count == 0 {
		return nil
	}

	result := make([]int, cb.count)
	i := cb.head
	for j := 0; j < cb.count; j++ {
		result[j] = cb.data[i]
		i = (i + 1) % cb.size
	}
	cb.head = 0
	cb.tail = 0
	cb.count = 0

	return result
}

// Источник данных из консоли
func source(out chan<- int, done chan bool) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data := scanner.Text()
		if data == "exit" {
			fmt.Println("The program has finished")
			close(done)
			return
		}
		st, err := strconv.Atoi(data)
		if err != nil {
			fmt.Println("There are not integers in input!")
			continue
		}
		out <- st
	}
}

func negativeFilter(input <-chan int, output chan<- int, done chan bool) {
	for {
		select {
		case data := <-input:
			if data >= 0 {
				output <- data
			}
		case <-done:
			return
		}
	}
}

func divFilter(input <-chan int, output chan<- int, done chan bool) {
	for {
		select {
		case data := <-input:
			if data%3 == 0 {
				output <- data
			}
		case <-done:
			return
		}
	}
}

func bufferData(input <-chan int, output chan<- []int, bufferSize int, interval time.Duration) {
	cb := NewCircularBuffer(bufferSize)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case num := <-input:
			cb.Push(num)
		case <-ticker.C:
			data := cb.Flush()
			if len(data) > 0 {
				output <- data
			}
		}
	}
}

func consumer(input <-chan []int, done chan bool) {
	for {
		select {
		case data := <-input:
			fmt.Println("Proccesed data: ", data)
		case <-done:
			return
		}
	}
}

func main() {
	input := make(chan int)
	done := make(chan bool)
	negativeChan := make(chan int)
	divChan := make(chan int)
	bufferedData := make(chan []int)
	const buffersize = 20
	const interval = time.Millisecond * 50

	go source(input, done)
	go negativeFilter(input, negativeChan, done)
	go divFilter(negativeChan, divChan, done)
	go bufferData(divChan, bufferedData, buffersize, interval)
	consumer(bufferedData, done)
}
