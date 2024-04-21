package utils

import "fmt"

type Pair[First any, Second any] struct {
	First  First
	Second Second
}

func (p *Pair[First, Second]) String() string {
	return fmt.Sprintf("(%v, %v)", p.First, p.Second)
}

func (p *Pair[First, Second]) Decompose() (First, Second) {
	return p.First, p.Second
}

func MakePair[First any, Second any](first First, second Second) Pair[First, Second] {
	return Pair[First, Second]{
		First:  first,
		Second: second,
	}
}
