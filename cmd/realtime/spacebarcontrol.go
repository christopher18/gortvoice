package main

import (
	"os"
	"realtime/pkg/audioinput"
	"time"

	"github.com/charmbracelet/log"

	"sync"

	"github.com/eiannone/keyboard"
)

func checkExit(key keyboard.Key) {
	if key == keyboard.KeyEsc {
		keyboard.Close()
		os.Exit(0)
	}
}

type KeyEvent struct {
	timestamp time.Time
	char      rune
	key       keyboard.Key
}

func UnmuteOnSpacebar(audioInput *audioinput.StreamHandler) {
	// Setup keyboard events
	if err := keyboard.Open(); err != nil {
		log.Fatalf("Failed to initialize keyboard: %v", err)
	}
	defer keyboard.Close()

	// Create mutex and latest event
	var mu sync.Mutex
	var latestEvent *KeyEvent

	// Start keyboard monitoring goroutine
	go func() {
		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				log.Printf("Error reading keyboard: %v", err)
				continue
			}
			mu.Lock()
			latestEvent = &KeyEvent{
				timestamp: time.Now(),
				char:      char,
				key:       key,
			}
			mu.Unlock()
		}
	}()

	for {
		// Wait for initial space press
		mu.Lock()
		event := latestEvent
		mu.Unlock()

		if event != nil {
			checkExit(event.key)
			if event.key == keyboard.KeySpace {
				log.Debug("Listening for audio input")
				audioInput.Unmute()
				lastEventTime := event.timestamp

				// Clear the latest event so we don't reprocess it
				mu.Lock()
				latestEvent = nil
				mu.Unlock()
			spaceLoop:
				for {
					timer := time.NewTimer(500 * time.Millisecond)
					<-timer.C

					// Check if we've had a more recent space event
					mu.Lock()
					latest := latestEvent
					mu.Unlock()

					if latest != nil &&
						latest.key == keyboard.KeySpace &&
						latest.timestamp.After(lastEventTime) {
						// Update last event time and continue waiting
						lastEventTime = latest.timestamp
						// Clear the event so we don't reprocess it
						mu.Lock()
						latestEvent = nil
						mu.Unlock()
						continue
					}

					// No recent space event, exit loop
					log.Debug("Audio input is muted")
					audioInput.Mute()
					break spaceLoop
				}
			}
		}
		time.Sleep(10 * time.Millisecond) // Small sleep to prevent CPU spinning
	}
}
