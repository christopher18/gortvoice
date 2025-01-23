package openairealtime

// session.update
type SessionUpdate struct {
	EventID string `json:"event_id"`
	Type    string `json:"type"`
	Session struct {
		Modalities              []string `json:"modalities"`
		Instructions            string   `json:"instructions"`
		Voice                   string   `json:"voice"`
		InputAudioFormat        string   `json:"input_audio_format"`
		OutputAudioFormat       string   `json:"output_audio_format"`
		InputAudioTranscription struct {
			Model string `json:"model"`
		} `json:"input_audio_transcription"`
		TurnDetection struct {
			Type              string  `json:"type"`
			Threshold         float64 `json:"threshold"`
			PrefixPaddingMS   int     `json:"prefix_padding_ms"`
			SilenceDurationMS int     `json:"silence_duration_ms"`
			CreateResponse    bool    `json:"create_response"`
		} `json:"turn_detection"`
		ToolChoice              string  `json:"tool_choice"`
		Temperature             float64 `json:"temperature"`
		MaxResponseOutputTokens string  `json:"max_response_output_tokens"`
	} `json:"session"`
}

// input_audio_buffer.append
type InputAudioBufferAppend struct {
	EventID string `json:"event_id"`
	Type    string `json:"type"`
	Audio   string `json:"audio"` // base64 encoded audio
}

// response.audio.delta
type ResponseAudioDelta struct {
	EventID      string `json:"event_id"`
	Type         string `json:"type"`
	ResponseID   string `json:"response_id"`
	ItemID       string `json:"item_id"`
	OutputIndex  int    `json:"output_index"`
	ContentIndex int    `json:"content_index"`
	Delta        string `json:"delta"` // base64 encoded audio
}

type Config struct {
	APIKey string
}
