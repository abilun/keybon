package results

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	KeysPressedTotal   int
	KeysPressedCorrect int
	Accuracy           float32
	Duration           time.Duration
	WPM                float64
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			cmd = func() tea.Msg {
				return BackMsg{}
			}
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	lines := []string{
		"Results",
		"",
		fmt.Sprintf("Keys pressed: %d", m.KeysPressedTotal),
		fmt.Sprintf("Keys pressed correct: %d", m.KeysPressedCorrect),
		fmt.Sprintf("Accuracy: %.2f%%", m.Accuracy),
		fmt.Sprintf("Duration: %.2f seconds", m.Duration.Seconds()),
		"",
		fmt.Sprintf("WPM: %.2f", m.WPM),
	}

	return lipgloss.JoinVertical(lipgloss.Center, lines...)
}

func (m *Model) Reset() {
	*m = Model{}
}
