package calculator

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// operation adapts a unary or binary function to a list of operands.
type operation struct {
	arity int
	apply func(operands []float64) (float64, error)
}

// operations is the registry behind Calculate. Adding an operation means
// adding an entry here.
var operations = map[string]operation{
	"add":        binary(Add),
	"subtract":   binary(Subtract),
	"multiply":   binary(Multiply),
	"divide":     binary(Divide),
	"power":      binary(Power),
	"sqrt":       unary(Sqrt),
	"percentage": binary(Percentage),
}

// supported lists the operation names in a stable order for error messages.
var supported = strings.Join(slices.Sorted(maps.Keys(operations)), ", ")

// Calculate applies the named operation to operands. The number of operands
// must match the operation: one for "sqrt" and two for the others.
func Calculate(name string, operands []float64) (float64, error) {
	op, ok := operations[name]
	if !ok {
		return 0, fmt.Errorf("%w %q; supported operations: %s", ErrUnknownOperation, name, supported)
	}
	if len(operands) != op.arity {
		noun := "operands"
		if op.arity == 1 {
			noun = "operand"
		}
		return 0, fmt.Errorf("%w: %s expects %d %s, got %d", ErrInvalidOperand, name, op.arity, noun, len(operands))
	}
	return op.apply(operands)
}

func unary(fn func(x float64) (float64, error)) operation {
	return operation{arity: 1, apply: func(x []float64) (float64, error) { return fn(x[0]) }}
}

func binary(fn func(a, b float64) (float64, error)) operation {
	return operation{arity: 2, apply: func(x []float64) (float64, error) { return fn(x[0], x[1]) }}
}
