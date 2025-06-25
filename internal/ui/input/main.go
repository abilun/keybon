package input

import (
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

const (
	defaultWidth  = 100
	defaultHeight = 5
)

type Model struct {
	expectedText []rune
	typedText    []rune

	height int
	width  int

	Cursor cursor.Model

	pos   int
	focus bool

	CorrectStyle lipgloss.Style
	WrongStyle   lipgloss.Style
	PendingStyle lipgloss.Style
	CursorStyle  lipgloss.Style
}

// New() function creates a new Model with predefined styles.
func New() Model {
	c := cursor.New()
	c.SetMode(cursor.CursorBlink)

	return Model{
		expectedText: []rune(""),
		pos:          0,
		Cursor:       c,
		width:        defaultWidth,
		height:       defaultHeight,

		CorrectStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		WrongStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Strikethrough(true),
		PendingStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	}
}

// Reset() function resets cursor position and typed text.
func (m *Model) Reset() {
	m.pos = 0
	m.typedText = []rune("")
}

// SetSize() function sets the width and height of the model.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetExpectedText() function sets the expected text for the model.
func (m *Model) SetExpectedText(expected string) {
	// todo: sanitize
	m.expectedText = []rune(expected)
}

// GetExpectedText() function gets the expected text for the model.
func (m *Model) GetExpectedText() string {
	return string(m.expectedText)
}

// GetTypedText() function gets the current text for the model.
func (m *Model) GetTypedText() string {
	return string(m.typedText)
}

// SetTypedText() function sets the current text for the model.
func (m *Model) SetTypedText(current string) {
	m.typedText = []rune(current)
}

// Focus() function sets the focus to the model.
func (m *Model) Focus() tea.Cmd {
	m.focus = true
	return m.Cursor.Focus()
}

// Blur() function removes the focus from the model.
func (m *Model) Blur() {
	m.focus = false
	m.Cursor.Blur()
}

// Focused() function returns true if the model is focused.
func (m Model) Focused() bool {
	return m.focus
}

// When starts, cursor no blinking until I move cursor
// Init() function initializes the model.
func (m Model) Init() tea.Cmd {
	// m.cursor.BlinkCmd() ?
	return nil
}

// SetCursor() function sets the cursor position.
func (m *Model) SetCursor(pos int) {
	m.pos = clamp(pos, 0, len(m.typedText))
}

// Position() function returns the cursor position.
func (m Model) Position() int {
	return m.pos
}

// CursorLeft() function moves the cursor to the left.
func (m *Model) CursorLeft() {
	if m.pos > 0 {
		m.pos--
	}
}

// CursorRight() function moves the cursor to the right.
func (m *Model) CursorRight() {
	if m.pos < len(m.typedText) {
		m.pos++
	}
}

// CursorEnd() function moves the cursor to the end of the text.
func (m *Model) CursorEnd() {
	m.SetCursor(len(m.typedText))
}

// AtEnd() function returns true if the cursor is at the end of the text.
func (m *Model) AtEnd() bool {
	return m.pos == len(m.expectedText)
}

// Update() function updates the model based on the message.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Let's remember where the position of the cursor currently is so that if
	// the cursor position changes, we can reset the blink.
	var cmds []tea.Cmd
	var cmd tea.Cmd

	oldPos := m.pos

	switch msg := msg.(type) {
	case tea.KeyMsg:
		isCorrect := false
		isBack := false

		switch {
		case !m.Focused():
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+w"))):
			isBack = true
			m.deleteWordBackward()
		case key.Matches(msg, key.NewBinding(key.WithKeys("backspace"))):
			isBack = true
			m.deleteRune()
		case key.Matches(msg, key.NewBinding(key.WithKeys("left"))):
			m.CursorLeft()
		case key.Matches(msg, key.NewBinding(key.WithKeys("right"))):
			m.CursorRight()
			// Typing: compare against target *before* inserting
		case msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace:
			for _, r := range msg.Runes {
				if m.pos < len(m.expectedText) {
					expected := m.expectedText[m.pos]
					m.insertRune(r)
					if r == expected {
						isCorrect = true
					}
				}
			}
		}

		cmd = func() tea.Msg {
			return KeystrokeProcessedMsg{
				TypedChar:   msg.Runes,
				IsCorrect:   isCorrect,
				IsBackspace: isBack,
				Position:    m.pos,
				Timestamp:   time.Now(),
			}
		}
		cmds = append(cmds, cmd)
	}

	m.Cursor, cmd = m.Cursor.Update(msg)
	cmds = append(cmds, cmd)

	if oldPos != m.pos && m.Cursor.Mode() == cursor.CursorBlink {
		m.Cursor.Blink = false
		cmds = append(cmds, m.Cursor.BlinkCmd())
	}

	if m.AtEnd() {
		typed := string(m.typedText)
		cmd = func() tea.Msg {
			return InputCompleteMsg{
				TypedText: typed,
			}
		}
		m.Reset()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// insertRune() inserts a rune at the cursor position.
func (m *Model) insertRune(r rune) {
	if r == 0 {
		return
	}
	before := m.typedText[:m.pos]
	after := m.typedText[m.pos:]
	m.typedText = append(before, append([]rune{r}, after...)...)
	m.pos++
}

// deleteRune() deletes a rune at the cursor position.
func (m *Model) deleteRune() {
	if m.pos <= 0 || len(m.typedText) == 0 {
		return
	}
	m.typedText = append(m.typedText[:m.pos-1], m.typedText[m.pos:]...)
	m.pos--
}

// deleteWordBackward from textinput is also an option
// just wanted to try to make it from scratch

// deleteWord() deletes a word backward from the cursor position
func (m *Model) deleteWordBackward() {
	if m.pos == 0 {
		return
	}
	oldPos := m.pos
	start := oldPos
	// skip spaces
	for start > 0 && unicode.IsSpace(m.expectedText[start-1]) {
		start--
	}
	// find start of word
	for start > 0 && !unicode.IsSpace(m.expectedText[start-1]) {
		start--
	}
	// start now points to the start of the word
	m.typedText = append(m.typedText[:start], m.typedText[oldPos:]...)
	m.pos = start
}

// View() returns the view of the model.
func (m Model) View() string {
	// TODO: add viewport
	var (
		lines        []string
		currentLine  []string
		currentWidth int
		cursorPlaced bool
		wordBuffer   []rune
		startPos     int
	)

	// Helper to flush wordBuffer as a chunk (word + space)
	flush := func() {
		if len(wordBuffer) == 0 {
			return
		}
		var chunk strings.Builder
		chunkWidth := 0

		for i, r := range wordBuffer {
			style := m.PendingStyle
			pos := startPos + i
			if pos < len(m.typedText) {
				if m.typedText[pos] == r {
					style = m.CorrectStyle
				} else {
					style = m.WrongStyle
				}
			}

			styled := style.Render(string(r))
			if pos == m.pos && !cursorPlaced {
				m.Cursor.SetChar(string(r))
				styled = m.Cursor.View()
				cursorPlaced = true
			}

			chunk.WriteString(styled)
			chunkWidth += runewidth.RuneWidth(r)
		}

		// If chunk won't fit, flush line and start new one
		if currentWidth+chunkWidth > m.width {
			lines = append(lines, strings.Join(currentLine, ""))
			currentLine = nil
			currentWidth = 0
		}

		currentLine = append(currentLine, chunk.String())
		currentWidth += chunkWidth
		startPos += len(wordBuffer)
		wordBuffer = nil
	}

	for i, r := range m.expectedText {
		wordBuffer = append(wordBuffer, r)

		// if space or last char, flush buffer
		if unicode.IsSpace(r) || i == len(m.expectedText)-1 {
			flush()
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, ""))
	}

	// Ensure exactly m.height lines
	if len(lines) < m.height {
		for len(lines) < m.height {
			lines = append(lines, "")
		}
	} else if len(lines) > m.height {
		lines = lines[:m.height]
	}

	// ANSI-aware truncate and pad to exactly m.width
	for i := range lines {
		lineWidth := ansi.StringWidth(lines[i])
		if lineWidth < m.width {
			lines[i] += strings.Repeat(" ", m.width-lineWidth)
		} else {
			lines[i] = ansi.Truncate(lines[i], m.width, "")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
