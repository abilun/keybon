package main

import (
	"bytes"
	_ "embed"
	"io"
	"log"
	"os"

	"github.com/abilun/keybon/internal/generator/dumb"
	"github.com/abilun/keybon/internal/ui"
	"github.com/alecthomas/kong"
)

var CLI struct {
	File   string `help:"File to read words from" short:"f" long:"file"`
	Length int    `help:"Number of words to generate" short:"l" long:"length" default:"10"`
}

//go:embed assets/english200.txt
var english200 []byte

func main() {
	kong.Parse(&CLI)
	var inputReader io.Reader = bytes.NewReader(english200)

	if CLI.File != "" {
		file, err := os.Open(CLI.File)
		if err != nil {
			log.Fatalf("failed to open file %q: %v", CLI.File, err)
		}
		defer file.Close()
		inputReader = file
	}

	generator := dumb.New()
	generator.Fill(inputReader)

	// words := make([]string, 0, CLI.Length)
	// for i := 0; i < CLI.Length; i++ {
	// 	word, err := generator.Next()
	// 	if err != nil {
	// 		log.Fatalf("generation failed: %v", err)
	// 	}
	// 	words = append(words, word)
	// }
	// text := strings.Join(words, " ")

	if err := ui.StartMainScreen(generator, CLI.Length); err != nil {
		log.Fatalf("TUI failed: %v", err)
	}
}
