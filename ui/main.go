package ui

import (
	"keybon/ui/input"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	borderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 0)
)

type mainScreen struct {
	input input.Model
}

func (m mainScreen) Init() tea.Cmd {
	return m.input.Init()
}

func (m mainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m mainScreen) View() string {
	b := strings.Builder{}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(m.input.View()))
	b.WriteString("\n")

	return b.String()
}

func StartMainScreen(text string) error {
	inp := input.New()
	inp.SetTarget(text)
	inp.Focus()
	inp.SetSize(50, 5)

	ms := mainScreen{input: inp}

	p := tea.NewProgram(
		ms,
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
