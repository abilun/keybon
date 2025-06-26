package ngram

import "math/rand"

// WeightedChoice returns a random key from the map,
// with the probability of each key being proportional to its value.
func WeightedChoice(m map[string]int) string {
	total := 0
	for _, count := range m {
		total += count
	}
	r := rand.Intn(total)
	for word, count := range m {
		if r < count {
			return word
		}
		r -= count
	}
	for word := range m {
		return word
	}
	return ""
}

// RandomChoice returns a random key from the map.
func RandomChoice(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys[rand.Intn(len(keys))]
}
