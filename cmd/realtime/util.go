package main

import (
	"bytes"
	"encoding/binary"
	"os"

	"github.com/charmbracelet/log"
)

// WAV header structure
type wavHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32  // 36 + data size
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32  // 16 for PCM
	AudioFormat   uint16  // 1 for PCM
	NumChannels   uint16  // 1 for mono
	SampleRate    uint32  // 16000
	ByteRate      uint32  // SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 16
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32  // data size
}

func createWavHeader(dataSize uint32) wavHeader {
	return wavHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + dataSize,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1,
		NumChannels:   1,
		SampleRate:    16000,
		ByteRate:      16000 * 1 * 16 / 8,
		BlockAlign:    1 * 16 / 8,
		BitsPerSample: 16,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: dataSize,
	}
}

func int16ToLittleEndian(data []int16) []byte {
	buf := bytes.NewBuffer(nil)
	for _, v := range data {
		binary.Write(buf, binary.LittleEndian, v)
	}
	return buf.Bytes()
}

func appendToWaveFile(filename string, data []byte) {
	// If file doesn't exist, create it with WAV header
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		header := createWavHeader(0) // Initial data size is 0
		binary.Write(file, binary.LittleEndian, header)
		file.Close()
	}

	// Open file for reading and writing
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Update data size in header
	var header wavHeader
	binary.Read(file, binary.LittleEndian, &header)
	header.ChunkSize += uint32(len(data))
	header.Subchunk2Size += uint32(len(data))

	// Write updated header
	file.Seek(0, 0)
	binary.Write(file, binary.LittleEndian, header)

	// Append new data at the end
	file.Seek(0, 2) // Seek to end of file
	_, err = file.Write(data)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}
}
