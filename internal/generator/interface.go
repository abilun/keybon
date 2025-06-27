package generator

type Generator interface {
	Next() (string, error)
}
