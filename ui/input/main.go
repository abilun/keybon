package input

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultWidth  = 100
	defaultHeight = 5
)

type Model struct {
	target  []rune
	current []rune

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
	// c.SetMode(cursor.CursorStatic) // or CursorBlink
	c.SetMode(cursor.CursorBlink)

	return Model{
		target: []rune(""),
		pos:    0,
		Cursor: c,
		width:  defaultWidth,
		height: defaultHeight,

		CorrectStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		WrongStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Strikethrough(true),
		PendingStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	}
}

// SetSize() function sets the width and height of the model.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetTarget() function sets the target text for the model.
func (m *Model) SetTarget(target string) {
	// todo: sanitize
	m.target = []rune(target)
}

// SetCurrent() function sets the current text for the model.
func (m *Model) SetCurrent(current string) {
	m.current = []rune(current)
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
	m.pos = pos
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
	if m.pos < len(m.current) {
		m.pos++
	}
}

// CursorEnd() function moves the cursor to the end of the text.
func (m *Model) CursorEnd() {
	m.SetCursor(len(m.current))
}

// AtEnd() function returns true if the cursor is at the end of the text.
func (m *Model) AtEnd() bool {
	return m.pos == len(m.current)
}

// Update() function updates the model based on the message.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// Let's remember where the position of the cursor currently is so that if
	// the cursor position changes, we can reset the blink.
	oldPos := m.pos

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case !m.Focused():
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+w"))):
			m.deleteWordBackward()
			// m.deleteWord()
		case key.Matches(msg, key.NewBinding(key.WithKeys("backspace"))):
			m.deleteRune()
		case key.Matches(msg, key.NewBinding(key.WithKeys("left"))):
			m.CursorLeft()
		case key.Matches(msg, key.NewBinding(key.WithKeys("right"))):
			m.CursorRight()
		case msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace:
			for _, r := range msg.Runes {
				if m.pos < len(m.target) {
					m.insertRune(r)
				}
			}
		}

	case tea.WindowSizeMsg:
		// Ignore for now
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.Cursor, cmd = m.Cursor.Update(msg)
	cmds = append(cmds, cmd)

	if oldPos != m.pos && m.Cursor.Mode() == cursor.CursorBlink {
		m.Cursor.Blink = false
		cmds = append(cmds, m.Cursor.BlinkCmd())
	}

	return m, tea.Batch(cmds...)
}

// insertRune() inserts a rune at the cursor position.
func (m *Model) insertRune(r rune) {
	if r == 0 {
		return
	}
	before := m.current[:m.pos]
	after := m.current[m.pos:]
	m.current = append(before, append([]rune{r}, after...)...)
	m.pos++
}

// deleteRune() deletes a rune at the cursor position.
func (m *Model) deleteRune() {
	if m.pos <= 0 || len(m.current) == 0 {
		return
	}
	m.current = append(m.current[:m.pos-1], m.current[m.pos:]...)
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
	for start > 0 && unicode.IsSpace(m.current[start-1]) {
		start--
	}
	// find start of word
	for start > 0 && !unicode.IsSpace(m.current[start-1]) {
		start--
	}
	// start now points to the start of the word
	m.current = append(m.current[:start], m.current[oldPos:]...)
	m.pos = start
}

// View() returns the view of the model.
func (m Model) View() string {
	height, width := m.height, m.width

	var lines []string
	var line strings.Builder
	cursorRendered := false

	currentLineWidth := 0
	currentLineCount := 0
	i := 0

	// Tokenize as words
	for i < max(len(m.current), len(m.target)) && currentLineCount < height {
		var (
			// wordRunes    []rune
			wordStr      string
			wordWidth    int
			wordIsCursor bool
			wordBuilder  strings.Builder
		)

		// Collect one "word" — space or sequence of non-spaces
		// start := i
		for ; i < max(len(m.current), len(m.target)); i++ {
			var r rune
			if i < len(m.current) {
				r = m.current[i]
			} else if i < len(m.target) {
				r = m.target[i]
			} else {
				break
			}

			// wordRunes = append(wordRunes, r)

			isCursor := m.Focused() && i == m.pos
			if isCursor {
				wordIsCursor = true
			}

			var styled string
			switch {
			case i < len(m.current) && i < len(m.target):
				if m.current[i] == m.target[i] {
					styled = m.CorrectStyle.Render(string(r))
				} else {
					styled = m.WrongStyle.Render(string(r))
				}
			case i < len(m.target):
				styled = m.PendingStyle.Render(string(r))
			default:
				styled = m.WrongStyle.Render(string(r))
			}

			if isCursor {
				m.Cursor.SetChar(string(r))
				wordBuilder.WriteString(m.Cursor.View())
			} else {
				wordBuilder.WriteString(styled)
			}

			wordWidth += lipgloss.Width(string(r))
			if unicode.IsSpace(r) {
				i++ // consume space
				break
			}
		}
		wordStr = wordBuilder.String()

		// If word won't fit on current line
		if currentLineWidth+wordWidth > width {
			// If word is longer than the line by itself, we must split it
			if currentLineWidth == 0 {
				line.WriteString(wordStr[:width]) // cut long word
				lines = append(lines, line.String())
				line.Reset()
				currentLineWidth = 0
				currentLineCount++
			} else {
				// Start new line
				lines = append(lines, line.String())
				line.Reset()
				currentLineWidth = 0
				currentLineCount++

				if currentLineCount >= height {
					break
				}

				line.WriteString(wordStr)
				currentLineWidth += wordWidth
			}
		} else {
			line.WriteString(wordStr)
			currentLineWidth += wordWidth
		}

		if wordIsCursor {
			cursorRendered = true
		}
	}

	if currentLineCount < height && line.Len() > 0 {
		lines = append(lines, line.String())
		currentLineCount++
	}

	// If cursor is at the very end and wasn't rendered
	if m.Focused() && !cursorRendered {
		if len(lines) == 0 {
			lines = append(lines, "")
		}
		lines[len(lines)-1] += "↵"
	}

	// Pad with empty lines if fixed height is specified
	if height > 0 {
		for len(lines) < height {
			lines = append(lines, "")
		}
		// Clip if overflowed
		if len(lines) > height {
			lines = lines[:height]
		}
	}

	return strings.Join(lines, "\n")

}
