# OpenAI Realtime Voice Assistant

A real-time voice assistant application that enables natural two-way conversations using OpenAI's real-time voice API.

## Overview

This project implements a voice interface to OpenAI's real-time API, allowing for natural back-and-forth conversations with an AI assistant. It handles audio input/output, real-time transcription, and streaming responses.

Key features:
- Real-time audio capture and playback
- Bi-directional WebSocket communication with OpenAI's API
- Voice activity detection for natural conversation flow (although currently you need to hold down the spacebar to talk)
- Configurable audio settings and assistant behavior
- Support for both text and audio modalities

## Project Structure

The project is organized into several key packages:

- `pkg/audioinput`: Handles audio capture from microphone using PortAudio
- `pkg/audiooutput`: Handles audio output from device using PortAudio
- `pkg/openairealtime`: Manages WebSocket communication with OpenAI's real-time API
- `cmd/realtime`: Contains the main application entry point and audio utilities

## Getting Started

### Prerequisites

- Go 1.19 or later
- PortAudio development libraries
- OpenAI API key with access to GPT-4 and real-time features

### Installation

1. Install PortAudio development libraries:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install portaudio19-dev

   # macOS
   brew install portaudio

   # Windows (using chocolatey)
   choco install portaudio
   ```

2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/openai-realtime-assistant
   cd openai-realtime-assistant
   ```

3. Install Go dependencies:
   ```bash
   go mod download
   ```

### Configuration

1. Sign up for OpenAI API access:
   - Visit [OpenAI's platform](https://platform.openai.com/signup)
   - Create an account or sign in
   - Navigate to API keys section
   - Create a new API key
   - Request access to GPT-4 and real-time features if needed

2. Set up environment variables:
   ```bash
   # Create .env file
   touch .env
   
   # Edit .env file with your API key
   OPENAI_API_KEY=your_api_key_here
   ```

### Running the Application

1. Build and run:
   ```bash
   # Using the provided script (loads .env automatically)
   ./run.sh
   
   # Or manually
   export $(cat .env | xargs)
   go run cmd/realtime/*.go
   ```

2. Controls:
   - **Hold SPACEBAR** to unmute your microphone and speak
   - **Release SPACEBAR** to stop (mic mutes after 500ms)
   - **Press ESC** to exit

3. The assistant will respond in real-time through your speakers.

### Troubleshooting

- If you encounter audio device issues, check the available input devices listed in the startup logs
- Ensure your microphone is properly connected and selected as the default input device
- Verify your OpenAI API key has the necessary permissions for real-time features

