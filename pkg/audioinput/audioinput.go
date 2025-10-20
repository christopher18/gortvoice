package audioinput

import (
	"errors"
	"time"

	"github.com/charmbracelet/log"

	"github.com/gordonklaus/portaudio"
)

// Config holds configuration for the audio capture
type Config struct {
	Channels        int // Number of audio channels (e.g., 1 for mono, 2 for stereo)
	SampleRate      int // Sample rate in Hz (e.g., 44100 for CD quality)
	FramesPerBuffer int // Number of frames per buffer
}

// StreamHandler wraps the PortAudio stream
type StreamHandler struct {
	stream *portaudio.Stream
	config Config
	muted  bool // Add muted flag
}

// Init initializes the PortAudio library
func Init(config Config) (*StreamHandler, error) {
	if err := portaudio.Initialize(); err != nil {
		log.Fatalf("Failed to initialize PortAudio: %v", err)
		return nil, err
	}

	// List available input devices
	devices, err := portaudio.Devices()
	if err != nil {
		log.Warnf("Failed to list available input devices: %v", err)
	} else {
		log.Infof("Available Input Devices:")
		for i, device := range devices {
			if device.MaxInputChannels > 0 {
				log.Infof("[%d] %s", i, device.Name)
			}
		}
	}

	sh := &StreamHandler{config: config}
	return sh, nil
}

func monitorChannel(chunkChan <-chan []byte) {
	log.Debug("Monitoring audio buffer usage")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Infof("Monitoring audio buffer usage tick")
		// Get current channel capacity usage
		used := len(chunkChan)
		capacity := cap(chunkChan)
		usagePercent := float64(used) / float64(capacity) * 100

		// Log based on severity
		switch {
		case usagePercent > 80:
			log.Warnf("Audio buffer nearly full! Usage: %.1f%% (%d/%d)",
				usagePercent, used, capacity)
		case usagePercent > 50:
			log.Infof("Audio buffer usage: %.1f%% (%d/%d)",
				usagePercent, used, capacity)
		default:
			log.Infof("Audio buffer usage: %.1f%% (%d/%d)",
				usagePercent, used, capacity)
		}
	}
}

// Listen starts capturing audio and returns a channel where PCM16 chunks are sent
func (sh *StreamHandler) Listen() (<-chan []byte, error) {
	if sh.config.Channels <= 0 {
		return nil, errors.New("invalid number of channels")
	}

	bufferCount := 50
	chunkChan := make(chan []byte, bufferCount)
	buffer := make([]int16, sh.config.FramesPerBuffer*sh.config.Channels)

	// Add monitoring goroutine
	go monitorChannel(chunkChan)

	stream, err := portaudio.OpenDefaultStream(
		sh.config.Channels, 0, float64(sh.config.SampleRate), len(buffer), func(input []int16) {
			if !sh.muted {
				chunkChan <- int16ToLittleEndian(input)
			}
		},
	)
	if err != nil {
		log.Fatalf("Failed to open default stream: %v", err)
		close(chunkChan)
		return nil, err
	}

	sh.stream = stream

	go func() {
		defer close(chunkChan)
		if err := stream.Start(); err != nil {
			// Handle stream start error
			log.Fatalf("Failed to start stream: %v", err)
		}
		// Stream will run until closed externally
		select {}
	}()

	return chunkChan, nil
}

// Close stops the stream and terminates PortAudio
func (sh *StreamHandler) Close() error {
	if sh.stream != nil {
		if err := sh.stream.Stop(); err != nil {
			return err
		}
		if err := sh.stream.Close(); err != nil {
			return err
		}
	}
	return portaudio.Terminate()
}

// Mute stops sending audio data without closing the stream
func (sh *StreamHandler) Mute() {
	sh.muted = true
	log.Debug("Audio input muted")
}

// Unmute resumes sending audio data
func (sh *StreamHandler) Unmute() {
	sh.muted = false
	log.Debug("Audio input unmuted")
}

// IsMuted returns the current mute status
func (sh *StreamHandler) IsMuted() bool {
	return sh.muted
}
