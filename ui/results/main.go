package results

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct{}

func (m Model) Init() tea.Cmd {
	return nil
}

type BackMsg struct{}

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
	sb := strings.Builder{}
	sb.WriteString("Results")
	return sb.String()
}
