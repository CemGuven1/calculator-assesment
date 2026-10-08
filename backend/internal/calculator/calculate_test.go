package calculator_test

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/CemGuven1/calculator-assesment/backend/internal/calculator"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		operands  []float64
		want      float64
		wantErr   error
	}{
		{name: "add", operation: "add", operands: []float64{2, 3}, want: 5},
		{name: "subtract", operation: "subtract", operands: []float64{5, 3}, want: 2},
		{name: "multiply", operation: "multiply", operands: []float64{4, 2.5}, want: 10},
		{name: "divide", operation: "divide", operands: []float64{10, 4}, want: 2.5},
		{name: "power", operation: "power", operands: []float64{2, 10}, want: 1024},
		{name: "sqrt takes one operand", operation: "sqrt", operands: []float64{16}, want: 4},
		{name: "percentage", operation: "percentage", operands: []float64{15, 200}, want: 30},
		{name: "operation errors are returned", operation: "divide", operands: []float64{1, 0}, wantErr: calculator.ErrDivisionByZero},

		{name: "unknown operation", operation: "modulo", operands: []float64{1, 2}, wantErr: calculator.ErrUnknownOperation},
		{name: "empty operation", operation: "", operands: []float64{1, 2}, wantErr: calculator.ErrUnknownOperation},
		{name: "names are case-sensitive", operation: "ADD", operands: []float64{1, 2}, wantErr: calculator.ErrUnknownOperation},

		{name: "nil operands", operation: "add", operands: nil, wantErr: calculator.ErrInvalidOperand},
		{name: "empty operands", operation: "add", operands: []float64{}, wantErr: calculator.ErrInvalidOperand},
		{name: "too few operands", operation: "add", operands: []float64{1}, wantErr: calculator.ErrInvalidOperand},
		{name: "too many operands", operation: "add", operands: []float64{1, 2, 3}, wantErr: calculator.ErrInvalidOperand},
		{name: "unary with no operands", operation: "sqrt", operands: []float64{}, wantErr: calculator.ErrInvalidOperand},
		{name: "unary with two operands", operation: "sqrt", operands: []float64{4, 9}, wantErr: calculator.ErrInvalidOperand},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := calculator.Calculate(tc.operation, tc.operands)
			assertResult(t, got, err, tc.want, tc.wantErr)
		})
	}
}

func TestArity(t *testing.T) {
	tests := []struct {
		operation string
		want      int
		wantErr   error
	}{
		{operation: "add", want: 2},
		{operation: "subtract", want: 2},
		{operation: "multiply", want: 2},
		{operation: "divide", want: 2},
		{operation: "power", want: 2},
		{operation: "sqrt", want: 1},
		{operation: "percentage", want: 2},
		{operation: "modulo", wantErr: calculator.ErrUnknownOperation},
		{operation: "", wantErr: calculator.ErrUnknownOperation},
	}
	for _, tc := range tests {
		t.Run(tc.operation, func(t *testing.T) {
			got, err := calculator.Arity(tc.operation)
			if !errors.Is(err, tc.wantErr) || got != tc.want {
				t.Errorf("Arity(%q) = (%d, %v), want (%d, %v)", tc.operation, got, err, tc.want, tc.wantErr)
			}
		})
	}
}

// Every operation must reject NaN and ±Inf in every operand position.
func TestCalculateRejectsNonFiniteOperands(t *testing.T) {
	operations := []struct {
		name  string
		arity int
	}{
		{"add", 2}, {"subtract", 2}, {"multiply", 2}, {"divide", 2},
		{"power", 2}, {"sqrt", 1}, {"percentage", 2},
	}
	nonFinite := []float64{math.NaN(), math.Inf(1), math.Inf(-1)}

	for _, op := range operations {
		for pos := range op.arity {
			for _, bad := range nonFinite {
				t.Run(fmt.Sprintf("%s/operand %d is %v", op.name, pos, bad), func(t *testing.T) {
					operands := make([]float64, op.arity)
					for i := range operands {
						operands[i] = 1
					}
					operands[pos] = bad

					got, err := calculator.Calculate(op.name, operands)
					assertResult(t, got, err, 0, calculator.ErrInvalidOperand)
				})
			}
		}
	}
}

// Messages are returned to API clients, so their wording is part of the contract.
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		call func() (float64, error)
		want string
	}{
		{
			name: "unknown operation",
			call: func() (float64, error) { return calculator.Calculate("modulo", []float64{1, 2}) },
			want: `unknown operation "modulo"; supported operations: add, divide, multiply, percentage, power, sqrt, subtract`,
		},
		{
			name: "too few operands",
			call: func() (float64, error) { return calculator.Calculate("add", []float64{1}) },
			want: "invalid operand: add expects 2 operands, got 1",
		},
		{
			name: "too many operands for a unary operation",
			call: func() (float64, error) { return calculator.Calculate("sqrt", []float64{4, 9}) },
			want: "invalid operand: sqrt expects 1 operand, got 2",
		},
		{
			name: "non-finite operand",
			call: func() (float64, error) { return calculator.Add(math.NaN(), 1) },
			want: "invalid operand: NaN is not a finite number",
		},
		{
			name: "division by zero",
			call: func() (float64, error) { return calculator.Divide(1, 0) },
			want: "division by zero",
		},
		{
			name: "zero to a negative power",
			call: func() (float64, error) { return calculator.Power(0, -1) },
			want: "division by zero: zero raised to a negative power",
		},
		{
			name: "square root of a negative",
			call: func() (float64, error) { return calculator.Sqrt(-4) },
			want: "result is not a real number: square root of a negative number",
		},
		{
			name: "negative base with a fractional exponent",
			call: func() (float64, error) { return calculator.Power(-8, 1.0/3) },
			want: "result is not a real number: negative base with a fractional exponent",
		},
		{
			name: "overflow",
			call: func() (float64, error) { return calculator.Multiply(math.MaxFloat64, 2) },
			want: "result is out of range",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.call()
			if err == nil {
				t.Fatalf("got no error, want %q", tc.want)
			}
			if err.Error() != tc.want {
				t.Errorf("message = %q, want %q", err.Error(), tc.want)
			}
		})
	}
}
