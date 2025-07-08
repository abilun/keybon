package ui

import (
	"strings"

	"github.com/abilun/keybon/internal/generator"
	"github.com/abilun/keybon/internal/typing"
	"github.com/abilun/keybon/internal/ui/input"
	"github.com/abilun/keybon/internal/ui/keyboard"
	"github.com/abilun/keybon/internal/ui/results"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type State int

const (
	mainView State = iota
	resultsView
)

type model struct {
	state State

	typingSession typing.TypingSession
	generator     generator.Generator
	wordsCount    int

	input         input.Model
	resultsScreen results.Model
	keyboard      keyboard.Model

	height int
	width  int
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.input.Init(),
		m.resultsScreen.Init(),
		m.keyboard.Init(),
		func() tea.Msg {
			return refreshWordsMsg{}
		},
	)
}

type refreshWordsMsg struct{}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case refreshWordsMsg:
		words := make([]string, 0, m.wordsCount)
		for i := 0; i < m.wordsCount; i++ {
			word, err := m.generator.Next()
			if err != nil {
				return m, tea.Quit
			}
			words = append(words, word)
		}
		text := strings.Join(words, " ")
		m.input.SetExpectedText(text)

	case input.InputCompleteMsg:
		m.typingSession.TypedText = msg.TypedText
		m.state = resultsView
		// TODO: worth setting somewhere else to decouple session from input
		m.typingSession.ExpectedText = m.input.GetExpectedText()
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

		cmds = append(cmds, func() tea.Msg {
			return refreshWordsMsg{}
		})

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case input.KeystrokeProcessedMsg:
		// Keystroke message is not a keystroke in a scope of typing session
		keystroke := typing.Keystroke{
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

		nextChar := m.input.NextChar()
		m.keyboard.NextKey(string(nextChar))
		m.keyboard, cmd = m.keyboard.Update(msg)

		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
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
		b.WriteString("\n")
		b.WriteString(m.keyboard.View())
		view = b.String()
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, view)
}

func New() model {
	input := input.New()
	input.Focus()
	keyboard, _ := keyboard.New(keyboard.En)

	return model{
		input:    input,
		state:    mainView,
		keyboard: keyboard,
	}
}

func StartMainScreen(gen generator.Generator, wordsCount int) error {
	ms := New()
	ms.generator = gen
	ms.wordsCount = wordsCount

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
	// 	slog.NewJSONHandler(
	// 		f,
	// 		&slog.HandlerOptions{
	// 			Level: slog.LevelDebug,
	// 		})))

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
