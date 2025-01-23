package audiooutput

import (
	"errors"
	"sync"
)

// Helper function to convert byte slice to int16 slice
func bytesToInt16(byteArray []byte) ([]int16, error) {
	if len(byteArray)%2 != 0 {
		return nil, errors.New("byte array length must be even for PCM16 data")
	}

	int16Array := make([]int16, len(byteArray)/2)
	for i := 0; i < len(byteArray); i += 2 {
		// Combine two bytes into an int16 (little-endian)
		int16Array[i/2] = int16(byteArray[i]) | int16(byteArray[i+1])<<8
	}

	return int16Array, nil
}

type CircularBuffer struct {
	buffer []int16
	size   int
	start  int
	end    int
	mu     sync.Mutex
}

// NewCircularBuffer creates a new circular buffer with the given size
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		buffer: make([]int16, size),
		size:   size,
	}
}

// Write adds data to the buffer
func (cb *CircularBuffer) Write(data []int16) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	for _, sample := range data {
		cb.buffer[cb.end] = sample
		cb.end = (cb.end + 1) % cb.size
		if cb.end == cb.start { // Overwrite if full
			cb.start = (cb.start + 1) % cb.size
		}
	}
}

// Read reads data from the buffer into the given slice
func (cb *CircularBuffer) Read(out []int16) int {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	count := 0
	for count < len(out) && cb.start != cb.end {
		out[count] = cb.buffer[cb.start]
		cb.start = (cb.start + 1) % cb.size
		count++
	}

	// Fill the remaining space with silence (zeros)
	for count < len(out) {
		out[count] = 0
		count++
	}

	return count
}
