package audiooutput

import (
	"errors"
	"fmt"
	"log"

	"github.com/gordonklaus/portaudio"
)

type Config struct {
	Channels        int
	SampleRate      int
	FramesPerBuffer int
}

type StreamHandler struct {
	config Config
	stream *portaudio.Stream
}

func Init(config Config) (*StreamHandler, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}

	return &StreamHandler{
		config: config,
	}, nil
}

func (sh *StreamHandler) Play(audioChan <-chan []byte) error {
	if sh.config.Channels <= 0 {
		return errors.New("invalid number of channels")
	}

	// Create a circular buffer with enough capacity to handle bursts of data
	circularBuffer := NewCircularBuffer(15000000)

	// Open PortAudio stream
	stream, err := portaudio.OpenDefaultStream(
		0, sh.config.Channels, // 0 input channels, N output channels
		float64(sh.config.SampleRate),
		sh.config.FramesPerBuffer, // Fixed size for PortAudio
		func(out []int16) {
			// Fill the output buffer from the circular buffer
			circularBuffer.Read(out)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to open audio stream: %w", err)
	}

	if err := stream.Start(); err != nil {
		return fmt.Errorf("failed to start audio stream: %w", err)
	}

	sh.stream = stream

	// Start playback loop
	go func() {
		defer stream.Close()
		for audioData := range audioChan {
			// Convert byte slice to int16 slice
			int16Data, err := bytesToInt16(audioData)
			if err != nil {
				log.Printf("Error converting byte data to int16: %v", err)
				continue
			}

			// Write data to the circular buffer
			circularBuffer.Write(int16Data)
		}
	}()

	return nil
}

func (sh *StreamHandler) Close() error {
	if sh.stream != nil {
		return sh.stream.Close()
	}
	return nil
}
