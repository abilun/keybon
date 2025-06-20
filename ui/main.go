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
			BorderForeground(lipgloss.Color("63"))
	greaterStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("63")).
			Align(lipgloss.Center)
)

type mainScreen struct {
	Input  input.Model
	height int
	width  int
}

func (m mainScreen) Init() tea.Cmd {
	return tea.Batch(
		m.Input.Init(),
	)
}

func (m mainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil

	case input.InputCompleteMsg:
		// TODO: results screen
		return m, tea.Quit
	}

	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m mainScreen) View() string {
	b := strings.Builder{}

	inputView := m.Input.View()
	inputViewBorder := borderStyle.Render(inputView)
	b.WriteString(greaterStyle.Width(lipgloss.Width(inputViewBorder)).Render("Keybon"))
	b.WriteString("\n")
	b.WriteString(inputViewBorder)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func New() mainScreen {
	input := input.New()
	input.Focus()

	return mainScreen{
		Input: input,
	}
}

func StartMainScreen(text string) error {
	ms := New()
	ms.Input.SetTarget(text)

	p := tea.NewProgram(
		ms,
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
