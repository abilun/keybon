package ui

import "time"

type TypingSession struct {
	Keystrokes []Keystroke
}

type TypingStats struct {
	KeysPressedTotal   int
	KeysPressedCorrect int
	Accuracy           float32
	FirstKeystroke     time.Time
	LastKeystroke      time.Time
	Duration           time.Duration
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
			// total
		}
		if k.IsCorrect {
			stats.KeysPressedCorrect++
		}

	}
	stats.Accuracy = float32(stats.KeysPressedCorrect) / float32(stats.KeysPressedTotal) * 100
	stats.FirstKeystroke = ts.Keystrokes[0].Timestamp
	stats.LastKeystroke = ts.Keystrokes[len(ts.Keystrokes)-1].Timestamp
	stats.Duration = stats.LastKeystroke.Sub(stats.FirstKeystroke)
	return stats
}
