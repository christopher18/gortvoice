package openairealtime

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
})

type OpenAIRealtimeClient struct {
	conn               *websocket.Conn
	audioOutput        chan<- []byte
	audioInput         <-chan []byte
	assistantIsTalking bool
}

// AttachAudioOutput attaches an audio output channel for assistant -> client communication
func (c *OpenAIRealtimeClient) AttachAudioOutput(output chan<- []byte) {
	c.audioOutput = output
}

// AttachAudioInput attaches an audio input channel for client -> assistant communication
func (c *OpenAIRealtimeClient) AttachAudioInput(input <-chan []byte) {
	c.audioInput = input
}

// GetOpenAIRealtimeClient initializes a new OpenAI Realtime client
func GetOpenAIRealtimeClient(config Config) (*OpenAIRealtimeClient, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Errorf("Error loading .env file: %v", err)
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	var apikey string
	if config.APIKey == "" {
		apikey = os.Getenv("OPENAI_API_KEY")
	} else {
		apikey = config.APIKey
	}

	if apikey == "" {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}

	// Open the key log file for writing
	keyLogFile, err := os.Create("keylogfile.log")
	if err != nil {
		log.Fatalf("Failed to create key log file: %v", err)
	}
	defer keyLogFile.Close()

	// Configure TLS with KeyLogWriter
	tlsConfig := &tls.Config{
		KeyLogWriter: keyLogFile,
	}

	url := "wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-12-17"
	dialer := websocket.Dialer{
		// HandshakeTimeout:  45 * time.Second,
		EnableCompression: true, // Try disabling compression if you're having issues
		ReadBufferSize:    1024 * 1024,
		WriteBufferSize:   1024 * 1024,
		TLSClientConfig:   tlsConfig,
	}
	headers := http.Header{
		"Authorization":            []string{"Bearer " + apikey},
		"OpenAI-Beta":              []string{"realtime=v1"},
		"Sec-WebSocket-Extensions": []string{"-permessage-deflate"},
	}

	conn, resp, err := dialer.Dial(url, headers)
	if err != nil {
		logger.Printf("WebSocket dial error: %v", err)
		if resp != nil {
			logger.Printf("HTTP Response: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	client := &OpenAIRealtimeClient{conn: conn}

	// go client.listenForEvents()

	return client, nil
}

func (c *OpenAIRealtimeClient) Start() error {
	if c.audioOutput == nil {
		return errors.New("audio output channel is not attached")
	}
	if c.audioInput == nil {
		return errors.New("audio input channel is not attached")
	}

	// Send initial session config
	if err := sendInitialSessionConfig(c.conn); err != nil {
		return fmt.Errorf("failed to send initial session config: %w", err)
	}

	// Start pinger
	// go c.startPinger()
	// Start listening for events from assistant
	go c.listenForEvents()
	// Start listening for audio input from client
	go c.listenForAudioInput()

	select {}
}

func (c *OpenAIRealtimeClient) listenForEvents() {
	defer c.conn.Close()

	// Set read deadline to detect stale connections
	c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))

	// Set ping handler to keep connection alive
	c.conn.SetPingHandler(func(appData string) error {
		// Log bytes in hex format for better debugging
		logger.Infof("Received ping bytes: %x", []byte(appData))
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))

		// Send pong with the same data we received
		return c.conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(25*time.Second))
	})

	// Set pong handler to log round-trip latency or extend the connection lifespan
	c.conn.SetPongHandler(func(appData string) error {
		logger.Printf("Received pong: %s", appData)
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))
		return nil
	})

	for {
		messageType, r, err := c.conn.NextReader()
		if err != nil {
			logger.Printf("Error reading raw frame: %v, time: %d", err, time.Now().UnixMilli())
			return
		}

		logger.Printf("Message Type: %d", messageType)
		message, err := io.ReadAll(r)
		if err != nil {
			logger.Printf("Error reading message: %v", err)
			logger.Printf("Raw Message: %s", string(message))
			return
		}

		// Reset read deadline after successful read
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))

		var data map[string]interface{}
		if err := json.Unmarshal(message, &data); err != nil {
			logger.Printf("Error unmarshalling message: %v, %s", err, string(message))
			continue
		}

		logger.Debugf("Received event: %s", prettyPrint(data))
		logger.Infof("Received event type: %s", data["type"])

		switch data["type"] {
		case "response.audio.delta":
			c.assistantIsTalking = true
			go handleResponseAudioDelta(c, message)
		case "response.created":
			c.assistantIsTalking = true
		case "response.audio.done":
			c.assistantIsTalking = false
		case "response.done":
			c.assistantIsTalking = false
		case "conversation.item.created":
			c.assistantIsTalking = true
		default:
			logger.Printf("Unhandled event type: %s", data["type"])
		}
	}
}

func (c *OpenAIRealtimeClient) listenForAudioInput() {
	for audio := range c.audioInput {
		if !c.assistantIsTalking {
			c.sendEvent(InputAudioBufferAppend{
				EventID: uuid.NewString(),
				Type:    "input_audio_buffer.append",
				Audio:   base64.StdEncoding.EncodeToString(audio),
			})
		}
	}
}

func (c *OpenAIRealtimeClient) sendEvent(event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshalling event: %w", err)
	}
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

func sendInitialSessionConfig(conn *websocket.Conn) error {
	sessionUpdate := SessionUpdate{
		EventID: uuid.NewString(),
		Type:    "session.update",
	}
	sessionUpdate.Session.Modalities = []string{"text", "audio"}
	sessionUpdate.Session.Instructions = initialPrompt
	sessionUpdate.Session.Voice = "ash"
	sessionUpdate.Session.InputAudioFormat = "pcm16"
	sessionUpdate.Session.OutputAudioFormat = "pcm16"
	sessionUpdate.Session.InputAudioTranscription.Model = "whisper-1"
	sessionUpdate.Session.TurnDetection = struct {
		Type              string  `json:"type"`
		Threshold         float64 `json:"threshold"`
		PrefixPaddingMS   int     `json:"prefix_padding_ms"`
		SilenceDurationMS int     `json:"silence_duration_ms"`
		CreateResponse    bool    `json:"create_response"`
	}{
		Type:              "server_vad",
		Threshold:         0.75,
		PrefixPaddingMS:   300,
		SilenceDurationMS: 500,
		CreateResponse:    true,
	}
	sessionUpdate.Session.ToolChoice = "auto"
	sessionUpdate.Session.Temperature = 0.8
	sessionUpdate.Session.MaxResponseOutputTokens = "inf"

	payload, err := json.Marshal(sessionUpdate)
	if err != nil {
		return fmt.Errorf("error marshalling session update: %w", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		logger.Errorf("Error sending session update: %v", err)
		return fmt.Errorf("error sending session update: %w", err)
	}

	logger.Printf("Connected to server. Sent initial session config.")
	return nil
}

func (c *OpenAIRealtimeClient) startPinger() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			logger.Printf("Failed to send ping: %v", err)
			return
		}
	}
}
