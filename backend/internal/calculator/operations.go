package calculator

import (
	"errors"
	"fmt"
	"math"
)

// Errors returned by this package. Most are wrapped with details about the
// specific failure, so match them with errors.Is.
var (
	// ErrUnknownOperation means the operation name is not supported.
	ErrUnknownOperation = errors.New("unknown operation")
	// ErrInvalidOperand means the number of operands is wrong for the
	// operation, or an operand is NaN or infinite.
	ErrInvalidOperand = errors.New("invalid operand")
	// ErrDivisionByZero means the operation divides by zero, which includes
	// raising zero to a negative power.
	ErrDivisionByZero = errors.New("division by zero")
	// ErrDomain means the result is not a real number, such as the square
	// root of a negative number.
	ErrDomain = errors.New("result is not a real number")
	// ErrOverflow means the result is too large in magnitude for a float64.
	ErrOverflow = errors.New("result is out of range")
)

// Add returns a + b.
func Add(a, b float64) (float64, error) {
	if err := checkOperands(a, b); err != nil {
		return 0, err
	}
	return checkResult(a + b)
}

// Subtract returns a - b.
func Subtract(a, b float64) (float64, error) {
	if err := checkOperands(a, b); err != nil {
		return 0, err
	}
	return checkResult(a - b)
}

// Multiply returns a × b.
func Multiply(a, b float64) (float64, error) {
	if err := checkOperands(a, b); err != nil {
		return 0, err
	}
	return checkResult(a * b)
}

// Divide returns a ÷ b.
func Divide(a, b float64) (float64, error) {
	if err := checkOperands(a, b); err != nil {
		return 0, err
	}
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return checkResult(a / b)
}

// Power returns base raised to exponent. Zero to the power of zero is 1.
//
// A negative base needs an integer exponent. Every non-integer float64 is a
// fraction with an even denominator, so a negative base raised to it never
// has a real result.
func Power(base, exponent float64) (float64, error) {
	if err := checkOperands(base, exponent); err != nil {
		return 0, err
	}
	if base == 0 && exponent < 0 {
		return 0, fmt.Errorf("%w: zero raised to a negative power", ErrDivisionByZero)
	}
	if base < 0 && exponent != math.Trunc(exponent) {
		return 0, fmt.Errorf("%w: negative base with a fractional exponent", ErrDomain)
	}
	return checkResult(math.Pow(base, exponent))
}

// Sqrt returns the square root of x.
func Sqrt(x float64) (float64, error) {
	if err := checkOperands(x); err != nil {
		return 0, err
	}
	if x < 0 {
		return 0, fmt.Errorf("%w: square root of a negative number", ErrDomain)
	}
	return checkResult(math.Sqrt(x))
}

// Percentage returns percent % of value, that is percent × value / 100, so
// Percentage(15, 200) is 30.
//
// Multiplying first keeps common cases exact (7% of 300 is 21 rather than
// 21.000000000000004). The cost is that a product beyond the float64 range
// overflows even when the final result would fit.
func Percentage(percent, value float64) (float64, error) {
	if err := checkOperands(percent, value); err != nil {
		return 0, err
	}
	return checkResult(percent * value / 100)
}

// checkOperands rejects NaN and ±Inf. JSON cannot encode them, but Go
// callers can pass them.
func checkOperands(operands ...float64) error {
	for _, x := range operands {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return fmt.Errorf("%w: %v is not a finite number", ErrInvalidOperand, x)
		}
	}
	return nil
}

// checkResult is the final guard on every result. It ensures nothing
// non-finite is returned, and normalizes -0 so it is never shown as "-0".
func checkResult(r float64) (float64, error) {
	switch {
	case math.IsNaN(r):
		return 0, ErrDomain
	case math.IsInf(r, 0):
		return 0, ErrOverflow
	case r == 0:
		return 0, nil // true for -0 as well
	default:
		return r, nil
	}
}
