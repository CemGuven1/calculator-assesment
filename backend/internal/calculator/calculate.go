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

// Arity returns the number of operands the named operation takes: one for
// "sqrt" and two for the others. It returns ErrUnknownOperation if there is
// no such operation.
func Arity(name string) (int, error) {
	op, ok := operations[name]
	if !ok {
		return 0, fmt.Errorf("%w %q; supported operations: %s", ErrUnknownOperation, name, supported)
	}
	return op.arity, nil
}

// Calculate applies the named operation to operands, whose number must match
// the operation's Arity.
func Calculate(name string, operands []float64) (float64, error) {
	arity, err := Arity(name)
	if err != nil {
		return 0, err
	}
	if len(operands) != arity {
		noun := "operands"
		if arity == 1 {
			noun = "operand"
		}
		return 0, fmt.Errorf("%w: %s expects %d %s, got %d", ErrInvalidOperand, name, arity, noun, len(operands))
	}
	return operations[name].apply(operands)
}

func unary(fn func(x float64) (float64, error)) operation {
	return operation{arity: 1, apply: func(x []float64) (float64, error) { return fn(x[0]) }}
}

func binary(fn func(a, b float64) (float64, error)) operation {
	return operation{arity: 2, apply: func(x []float64) (float64, error) { return fn(x[0], x[1]) }}
}
