package generator

import (
	"errors"
	"strings"

	"keybon/internal/ngram/model"
)

type NgramGenerator struct {
	*model.Model
	history  []string
	nextFunc func(map[string]int) string
}

// TODO: what if I first call New, then Import?
// New() function creates a new NgramGenerator with the given order.
func New(order int) *NgramGenerator {
	if order < 1 {
		panic("order must be greater than 0")
	}
	ng := &NgramGenerator{
		Model: &model.Model{
			Order: order,
			Data:  make(map[string]map[string]int),
		},
		history: make([]string, 0, order),
	}
	ng.NextFunc(WeightedChoice)

	return ng
}

func (ng *NgramGenerator) NextFunc(nextFunc func(map[string]int) string) {
	ng.nextFunc = nextFunc
}

// Start() function starts the model and sets
// the initial history to the first key in the model.
func (ng *NgramGenerator) Start() error {
	if len(ng.Model.Data) == 0 {
		return errors.New("model is empty")
	}
	for k := range ng.Model.Data {
		ng.history = strings.Split(k, " ")
		break
	}
	return nil
}

// Next() function returns the next word in the model based on the current history.
func (ng *NgramGenerator) Next() (string, error) {
	if ng.Model.IsEmpty() {
		return "", errors.New("model is empty")
	}
	if ng.history == nil {
		return "", errors.New("generator is not started")
	}

	key := strings.Join(ng.history, " ")
	nexts, ok := ng.Model.Data[key]
	if !ok || len(nexts) == 0 {
		return "", errors.New("no next words")
	}
	next := WeightedChoice(nexts)
	ng.history = append(ng.history[1:], next)
	return next, nil
}
