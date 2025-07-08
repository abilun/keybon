package keyboard

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	pressedKey string
	keyLayout  [][]string
	nextKey    string

	KeyStyle     lipgloss.Style
	NextKeyStyle lipgloss.Style
}

type Language int

const (
	En Language = iota
)

var (
	defaultKeyStyle = lipgloss.NewStyle().
			Width(5).
			Height(1).
			Align(lipgloss.Center, lipgloss.Center).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("255")).
			Foreground(lipgloss.Color("255"))

	defaultNextKeyStyle = defaultKeyStyle.
				Foreground(lipgloss.Color("82")).
				BorderForeground(lipgloss.Color("82")).
				Bold(true)
)

func (m *Model) NextKey(k string) {
	m.nextKey = k
}

func New(lang Language) (Model, error) {
	switch lang {
	case En:
		return Model{
			keyLayout: [][]string{
				{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p"},
				{"a", "s", "d", "f", "g", "h", "j", "k", "l", ";"},
				{"z", "x", "c", "v", "b", "n", "m", ",", ".", "/"},
			},
			KeyStyle:     defaultKeyStyle,
			NextKeyStyle: defaultNextKeyStyle,
		}, nil
	default:
		return Model{}, fmt.Errorf("unsupported language: %v", lang)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyRunes {
			keyStr := msg.String()
			if len(keyStr) == 1 {
				m.pressedKey = keyStr
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	var keyboardRows []string
	for _, row := range m.keyLayout {
		var keys []string
		for _, k := range row {
			var keyRender string
			if k == m.nextKey {
				keyRender = defaultNextKeyStyle.Render(strings.ToUpper(k))
			} else {
				keyRender = defaultKeyStyle.Render(strings.ToUpper(k))
			}
			keys = append(keys, keyRender)
		}

		keyboardRows = append(keyboardRows, lipgloss.JoinHorizontal(lipgloss.Top, keys...))
	}

	keyboard := lipgloss.JoinVertical(lipgloss.Left, keyboardRows...)
	s.WriteString(keyboard)

	return s.String()
}
