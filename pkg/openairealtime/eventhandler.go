package openairealtime

import (
	"encoding/base64"
	"encoding/json"

	"github.com/charmbracelet/log"
)

// handleResponseAudioDelta handles the response audio delta event
func handleResponseAudioDelta(client *OpenAIRealtimeClient, b []byte) {
	audioDelta := ResponseAudioDelta{}
	json.Unmarshal(b, &audioDelta)

	delta, err := base64.StdEncoding.DecodeString(audioDelta.Delta)
	if err != nil {
		log.Printf("Error decoding base64 audio delta: %v", err)
		return
	}

	client.audioOutput <- delta
}
