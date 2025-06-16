package model

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"keybon/scanner"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// TODO: check if fill, load, save consistent on order
// TODO: revisit fields & methods that should be exported

type Model struct {
	Order  int                       `json:"order"`
	Data   map[string]map[string]int `json:"data"`
	Hashes map[string]struct{}       `json:"hashes"`
}

// Add() function adds a word to the model with the given history.
func (m *Model) Add(history []string, word string) error {
	if len(history) == 0 {
		return errors.New("history is empty")
	}
	if len(history) > m.Order {
		return errors.New("history is longer than order")
	}

	key := strings.Join(history, " ")
	if _, ok := m.Data[key]; !ok {
		m.Data[key] = make(map[string]int)
	}
	m.Data[key][word]++

	return nil
}

// FillWithHash() function fills the model with data
// from a reader and stores the hash of the data.
func (m *Model) FillWithHash(r io.Reader) error {
	hasher := sha256.New()
	tee := io.TeeReader(r, hasher)
	err := m.Fill(tee)
	if err != nil {
		return err
	}

	sum := fmt.Sprintf("%x", hasher.Sum(nil))
	if m.Hashes == nil {
		m.Hashes = make(map[string]struct{})
	}

	if _, seen := m.Hashes[sum]; seen {
		return errors.New("duplicate content: already processed")
	}

	m.Hashes[sum] = struct{}{}
	return nil
}

// Fill() function fills the model with data from a reader.
func (m *Model) Fill(r io.Reader) error {
	scanner, err := scanner.New(r)
	if err != nil {
		return err
	}

	var history []string
	for scanner.Scan() {
		word := scanner.Text()
		if len(history) == m.Order {
			m.Add(history, word)
			history = history[1:]
		}
		history = append(history, word)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// EncodeJSON() function encodes the model to JSON.
func (m *Model) EncodeJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(m)
}

// DecodeJSON() function decodes the model from JSON.
func (m *Model) DecodeJSON(r io.Reader) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(m)
}

// CompressZstd() function compresses the model using Zstandard.
func (m *Model) CompressZstd(w io.Writer) error {
	encoder, err := zstd.NewWriter(w)
	if err != nil {
		return err
	}
	defer encoder.Close()
	return nil
}

// DecompressZstd() function decompresses the model using Zstandard.
func (m *Model) DecompressZstd(r io.Reader) error {
	decoder, err := zstd.NewReader(r)
	if err != nil {
		return err
	}
	defer decoder.Close()
	return nil
}

// Save() function writes the JSON encoded model
// to a writer using Zstandard compression.
func (m *Model) Save(w io.Writer) error {
	encoder, err := zstd.NewWriter(w)
	if err != nil {
		return err
	}
	defer encoder.Close()
	return m.EncodeJSON(encoder)
}

// Load() function reads the JSON encoded model
// from a reader using Zstandard decompression.
func (m *Model) Load(r io.Reader) error {
	decoder, err := zstd.NewReader(r)
	if err != nil {
		return err
	}
	defer decoder.Close()
	return m.DecodeJSON(decoder)
}

// IsEmpty() function returns true if the model is empty.
func (m *Model) IsEmpty() bool {
	return len(m.Data) == 0 || m.Data == nil
}

// Clear() function clears the model.
func (m *Model) Clear() {
	m.Data = make(map[string]map[string]int)
}
