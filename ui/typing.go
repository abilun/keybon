package ui

import (
	"strings"
	"time"
	"unicode"
)

type TypingSession struct {
	Keystrokes []Keystroke
	TargetText string
	TypedText  string
}

func (ts *TypingSession) Start(targetText string) {
	ts.TargetText = targetText
}

func (ts *TypingSession) Reset() {
	ts.Keystrokes = []Keystroke{}
	ts.TargetText = ""
}

func (ts *TypingSession) AddKeystroke(k Keystroke) {
	ts.Keystrokes = append(ts.Keystrokes, k)
}

type TypingStats struct {
	KeysPressedTotal   int
	KeysPressedCorrect int
	Accuracy           float32
	FirstKeystroke     time.Time
	LastKeystroke      time.Time
	Duration           time.Duration
	WPM                float64
}

type Keystroke struct {
	Position    int       `json:"position"`
	TypedChar   []rune    `json:"typed_char"`
	IsCorrect   bool      `json:"is_correct"`
	IsBackspace bool      `json:"is_backspace"`
	Timestamp   time.Time `json:"timestamp"`
}

func (ts TypingSession) Stats() TypingStats {
	stats := TypingStats{}

	for _, k := range ts.Keystrokes {
		if !k.IsBackspace {
			stats.KeysPressedTotal++
		}
		if k.IsCorrect {
			stats.KeysPressedCorrect++
		}
	}

	stats.Accuracy = float32(stats.KeysPressedCorrect) / float32(stats.KeysPressedTotal) * 100
	stats.FirstKeystroke = ts.Keystrokes[0].Timestamp
	stats.LastKeystroke = ts.Keystrokes[len(ts.Keystrokes)-1].Timestamp
	stats.Duration = stats.LastKeystroke.Sub(stats.FirstKeystroke)
	stats.WPM = ts.calculateSessionWPM()

	return stats
}

func (ts *TypingSession) calculateSessionWPM() float64 {
	if len(ts.Keystrokes) == 0 {
		return 0
	}

	expectedText := ts.TargetText
	counter := NewWPMCounter(expectedText, ts.TypedText)
	wordResults := counter.AnalyzeWords(ts.Keystrokes)

	// Get session timing
	if len(ts.Keystrokes) == 0 {
		return 0
	}

	startTime := ts.Keystrokes[0].Timestamp
	endTime := ts.Keystrokes[len(ts.Keystrokes)-1].Timestamp

	// Calculate WPM
	return counter.CalculateWPM(wordResults, startTime, endTime)
}

func (w *WPMCounter) CalculateWPM(wordResults []WordResult, startTime, endTime time.Time) float64 {
	correctChars := 0

	for i, result := range wordResults {
		if result.IsCorrect {
			correctChars += len(result.Word)

			// Add space after word except last word
			if i < len(wordResults)-1 {
				correctChars++ // space character
			}
		}
	}

	duration := endTime.Sub(startTime)
	minutes := duration.Minutes()

	if minutes == 0 {
		return 0
	}

	return float64(correctChars) / 5.0 / minutes
}

func (w *WPMCounter) analyzeWord(bound WordBoundary, finalText string, keystrokes []Keystroke) WordResult {
	// Get the word from final typed text
	var typedWord string
	if bound.StartPos < len(finalText) {
		endPos := min(bound.EndPos, len(finalText))
		typedWord = finalText[bound.StartPos:endPos]
	}

	// Check if word is correct
	isCorrect := typedWord == bound.Word

	// Check if word was corrected (had backspaces in word range)
	wasCorrected := w.hadCorrections(bound, keystrokes)

	return WordResult{
		Word:         bound.Word,
		StartPos:     bound.StartPos,
		EndPos:       bound.EndPos,
		IsCorrect:    isCorrect,
		WasCorrected: wasCorrected,
	}
}

// hadCorrections() checks if there were backspaces affecting this word
func (w *WPMCounter) hadCorrections(bound WordBoundary, keystrokes []Keystroke) bool {
	for _, ks := range keystrokes {
		if ks.IsBackspace && ks.Position >= bound.StartPos && ks.Position < bound.EndPos {
			return true
		}
	}
	return false
}

type WordResult struct {
	Word         string
	StartPos     int
	EndPos       int
	IsCorrect    bool
	WasCorrected bool
}

type WPMCounter struct {
	expectedText string
	typedText    string
	words        []string
	wordBounds   []WordBoundary
}

type WordBoundary struct {
	StartPos int
	EndPos   int
	Word     string
}

// NewWPMCounter() creates a counter for the given text
func NewWPMCounter(expectedText string, typedText string) *WPMCounter {
	counter := &WPMCounter{
		expectedText: expectedText,
		typedText:    typedText,
		words:        strings.Fields(expectedText),
	}
	counter.calculateWordBoundaries()
	return counter
}

// calculateWordBoundaries() finds start/end positions of each word
func (w *WPMCounter) calculateWordBoundaries() {
	pos := 0
	runes := []rune(w.expectedText)

	for _, word := range w.words {
		wordRunes := []rune(word)

		// Find word start position accounting for spaces
		for pos < len(runes) && unicode.IsSpace(runes[pos]) {
			pos++
		}

		if pos >= len(runes) {
			break
		}

		w.wordBounds = append(w.wordBounds, WordBoundary{
			StartPos: pos,
			EndPos:   pos + len(wordRunes),
			Word:     word,
		})

		pos += len(wordRunes)
	}
}

// AnalyzeWords() processes keystrokes and determines which words were correct
func (w *WPMCounter) AnalyzeWords(keystrokes []Keystroke) []WordResult {
	var results []WordResult

	// Build final typed text by simulating all keystrokes
	finalText := w.typedText

	for _, bound := range w.wordBounds {
		result := w.analyzeWord(bound, finalText, keystrokes)
		results = append(results, result)
	}

	return results
}
