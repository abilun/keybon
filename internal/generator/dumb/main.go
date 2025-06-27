package dumb

import (
	"errors"
	"io"
	"math/rand"

	"github.com/abilun/keybon/internal/scanner"
)

type Generator struct {
	words []string
}

func (g *Generator) Next() (string, error) {
	if len(g.words) == 0 {
		return "", errors.New("no words to generate")
	}
	return g.words[rand.Intn(len(g.words))], nil
}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Fill(r io.Reader) error {
	scanner, err := scanner.New(r)
	if err != nil {
		return err
	}

	for scanner.Scan() {
		word := scanner.Text()
		g.words = append(g.words, word)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
