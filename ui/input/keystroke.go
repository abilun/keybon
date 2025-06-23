package input

import (
	"time"
)

type KeystrokeProcessedMsg struct {
	Position    int
	TypedChar   []rune
	IsCorrect   bool
	IsBackspace bool
	Timestamp   time.Time
}
