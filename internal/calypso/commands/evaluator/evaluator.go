package evaluator

import (
	"fmt"
)

type Evaluator struct {
}

func New() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(program any) {
	fmt.Printf("%T\n", program)
}
