package scanner

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"unicode"
	"unicode/utf8"
)

// TODO: test config override
// Config probably should not be visible to the user
type TextScannerConfig struct {
	Lowercase bool
}

type TextScanner struct {
	Config TextScannerConfig
	*bufio.Scanner
}

func New(r io.Reader) (*TextScanner, error) {
	ts := &TextScanner{
		Config: TextScannerConfig{Lowercase: true},
	}

	ts.Scanner = bufio.NewScanner(r)
	ts.Scanner.Split(ts.scanWordsNormalized)
	return ts, nil
}

// NewWithConfig() function creates a new TextScanner with the given configuration.
func NewWithConfig(file *os.File, config TextScannerConfig) (*TextScanner, error) {
	ts := &TextScanner{
		Config: config,
	}

	ts.Scanner = bufio.NewScanner(file)
	ts.Scanner.Split(ts.scanWordsNormalized)
	return ts, nil
}

// SetConfig() function sets the configuration for the TextScanner.
func (ts *TextScanner) SetConfig(config TextScannerConfig) {
	ts.Config = config
}

func (ts *TextScanner) scanWordsNormalized(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := 0

	// Skip non-letter runes at the beginning
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if unicode.IsLetter(r) {
			break
		}
	}

	// Scan until a non-letter rune
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if !unicode.IsLetter(r) {
			// Return the word found
			var word []byte
			if ts.Config.Lowercase {
				word = bytes.ToLower(data[start:i])
			} else {
				word = data[start:i]
			}

			return i + width, word, nil
		}
	}

	// If at EOF and still data left, return the last word
	if atEOF && start < len(data) {
		word := data[start:]
		if ts.Config.Lowercase {
			word = bytes.ToLower(word)
		}
		return len(data), word, nil
	}

	// Need more data
	return start, nil, nil
}
