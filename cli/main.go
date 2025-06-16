package main

import (
	"flag"
	"fmt"
	"keybon/generator"
	"keybon/input"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	borderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2)
)

type model struct {
	input input.Model
}

func (m model) Init() tea.Cmd {
	return m.input.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) View() string {
	return "\n" + borderStyle.Render(m.input.View()) + "\n"
}

func main() {
	order := flag.Int("n", 2, "n-gram order (used only when creating a new model)")
	length := flag.Int("len", 25, "number of words to generate")
	file := flag.String("file", "", "input text file to train/update the model")
	loadPath := flag.String("load", "", "load model from this file")
	savePath := flag.String("save", "", "save model to this file")
	tui := flag.Bool("tui", true, "show TUI interface (default true)")

	flag.Parse()

	var ng *generator.NgramGenerator

	if *loadPath != "" {
		f, err := os.Open(*loadPath)
		if err != nil {
			log.Fatalf("failed to load model from %s: %v", *loadPath, err)
		}
		defer f.Close()

		ng = generator.New(*order)
		if err := ng.Load(f); err != nil {
			log.Fatalf("failed to decode model: %v", err)
		}
	} else {
		if *order < 1 {
			log.Fatalf("n-gram order must be at least 1")
		}
		ng = generator.New(*order)
	}

	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			log.Fatalf("failed to open input file: %v", err)
		}
		defer f.Close()

		if err := ng.FillWithHash(f); err != nil {
			log.Fatalf("failed to update model: %v", err)
		}
	}

	if *savePath != "" {
		f, err := os.Create(*savePath)
		if err != nil {
			log.Fatalf("failed to create save file: %v", err)
		}
		defer f.Close()

		if err := ng.Save(f); err != nil {
			log.Fatalf("failed to save model: %v", err)
		}
	}

	if err := ng.Start(); err != nil {
		log.Fatalf("model not ready: %v", err)
	}

	// Generate words
	words := make([]string, 0, *length)
	for i := 0; i < *length; i++ {
		word, err := ng.Next()
		if err != nil {
			log.Fatalf("generation failed: %v", err)
		}
		words = append(words, word)
	}
	text := strings.Join(words, " ")

	// TUI or plain output
	if *tui {
		inp := input.New()
		inp.SetTarget(text)
		inp.Focus()

		m := model{input: inp}
		p := tea.NewProgram(m, tea.WithAltScreen())
		if err := p.Start(); err != nil {
			log.Fatalf("TUI failed: %v", err)
		}
	} else {
		fmt.Println(text)
	}
}
