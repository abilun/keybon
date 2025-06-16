package input

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	target  []rune
	current []rune

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

		CorrectStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		WrongStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Strikethrough(true),
		PendingStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	}
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
	var b strings.Builder

	// Determine the maximum length to iterate over both current and target runes
	maxLen := max(len(m.current), len(m.target))

	for i := 0; i < maxLen; i++ {
		var (
			runeToShow rune
			style      lipgloss.Style
			isCursor   = m.Focused() && i == m.pos
		)

		switch {
		case i < len(m.current) && i < len(m.target):
			// User has typed this character — compare with target
			runeToShow = m.current[i]
			if runeToShow == m.target[i] {
				style = m.CorrectStyle
			} else {
				style = m.WrongStyle
			}

		case i < len(m.target):
			// User hasn't typed this character yet
			runeToShow = m.target[i]
			style = m.PendingStyle

		default:
			// Extra characters beyond target if allowed
			runeToShow = m.current[i]
			style = m.WrongStyle
		}

		// Apply cursor style if cursor is at this position
		if isCursor {
			m.Cursor.SetChar(string(runeToShow))
			b.WriteString(m.Cursor.View())
		} else {
			b.WriteString(style.Render(string(runeToShow)))
		}
	}

	// If the cursor is at the very end, show ↵ as a visual end-of-input marker
	if m.Focused() && m.pos == maxLen {
		b.WriteString("↵")
	}

	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
