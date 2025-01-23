package main

import (
	"os"
	"realtime/pkg/audioinput"
	"realtime/pkg/audiooutput"
	"realtime/pkg/openairealtime"

	"github.com/charmbracelet/log"
)

func main() {
	// Get an audio input stream
	audioInput, err := audioinput.Init(audioinput.Config{
		Channels:        1,
		SampleRate:      24000,
		FramesPerBuffer: 1920,
	})
	if err != nil {
		log.Fatalf("Failed to initialize audio input: %v", err)
	}

	inputchan, err := audioInput.Listen()
	if err != nil {
		log.Fatalf("Failed to initialize audio input: %v", err)
	}
	log.Printf("Listening for audio input")

	// Mute the audio input by default
	audioInput.Mute()

	go UnmuteOnSpacebar(audioInput)

	// Write audio to a file for testing
	// for {
	// 	audioData := <-inputchan
	// 	appendToWaveFile("audio.wav", int16ToLittleEndian(audioData))
	// }

	audioOutput, err := audiooutput.Init(audiooutput.Config{
		Channels:        1,
		SampleRate:      24000,
		FramesPerBuffer: 480,
	})
	if err != nil {
		log.Fatalf("Failed to initialize audio output: %v", err)
	}

	outputchan := make(chan []byte)

	err = audioOutput.Play(outputchan)
	if err != nil {
		log.Fatalf("Failed to play audio: %v", err)
	}

	// Get an OpenAI Realtime client
	openaiRealtime, err := openairealtime.GetOpenAIRealtimeClient(openairealtime.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize OpenAI Realtime client: %v", err)
	}

	openaiRealtime.AttachAudioInput(inputchan)
	openaiRealtime.AttachAudioOutput(outputchan)

	// Start the OpenAI Realtime client (blocking)
	err = openaiRealtime.Start()
	if err != nil {
		log.Fatalf("Failed to start OpenAI Realtime client: %v", err)
	}
}
