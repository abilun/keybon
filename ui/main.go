package ui

import (
	"keybon/ui/input"
	"keybon/ui/results"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	mainView State = iota
	resultsView
)

type mainScreen struct {
	state State

	typingSession TypingSession

	input         input.Model
	resultsScreen results.Model

	height int
	width  int
}

func (m mainScreen) Init() tea.Cmd {
	return tea.Batch(
		m.input.Init(),
		m.resultsScreen.Init(),
	)
}

func (m mainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case input.InputCompleteMsg:
		m.input.Reset()
		m.state = resultsView
		// TODO: worth setting somewhere else to decouple session from input
		m.typingSession.TargetText = m.input.GetTarget()
		stats := m.typingSession.Stats()

		m.resultsScreen = results.Model{
			KeysPressedTotal:   stats.KeysPressedTotal,
			KeysPressedCorrect: stats.KeysPressedCorrect,
			Accuracy:           stats.Accuracy,
			Duration:           stats.Duration,
			WPM:                stats.WPM,
		}

	case results.BackMsg:
		m.state = mainView
		m.resultsScreen.Reset()
		m.typingSession.Reset()

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case input.KeystrokeProcessedMsg:
		// Keystroke message is not a keystroke in a scope of typing session
		keystroke := Keystroke{
			Position:    msg.Position,
			TypedChar:   msg.TypedChar,
			IsCorrect:   msg.IsCorrect,
			IsBackspace: msg.IsBackspace,
			Timestamp:   msg.Timestamp,
		}
		m.typingSession.AddKeystroke(keystroke)
	}

	switch m.state {
	case resultsView:
		m.resultsScreen, cmd = m.resultsScreen.Update(msg)
		cmds = append(cmds, cmd)
	case mainView:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}
		}
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m mainScreen) View() string {
	b := strings.Builder{}
	var view string

	switch m.state {
	case resultsView:
		view = borderStyle.Render(m.resultsScreen.View())
	case mainView:
		inputViewBorder := borderStyle.Render(m.input.View())
		b.WriteString(greaterStyle.Width(lipgloss.Width(inputViewBorder)).Render("Keybon"))
		b.WriteString("\n")
		b.WriteString(inputViewBorder)
		view = b.String()
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, view)
}

func New() mainScreen {
	input := input.New()
	input.Focus()

	return mainScreen{
		input: input,
		state: mainView,
	}
}

func StartMainScreen(text string) error {
	ms := New()
	ms.input.SetTarget(text)

	p := tea.NewProgram(
		ms,
		tea.WithAltScreen(),
	)

	// f, err := tea.LogToFile("/tmp/debug.log", "debug")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer f.Close()

	// slog.SetDefault(slog.New(
	// 	slog.NewTextHandler(
	// 		f,
	// 		&slog.HandlerOptions{
	// 			Level: slog.LevelDebug,
	// 		})))

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
