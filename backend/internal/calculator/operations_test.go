package calculator_test

import (
	"errors"
	"math"
	"testing"

	"github.com/CemGuven1/calculator-assesment/backend/internal/calculator"
)

// negZero is IEEE 754 negative zero, which a Go constant cannot express.
var negZero = math.Copysign(0, -1)

type binaryCase struct {
	name    string
	a, b    float64
	want    float64
	wantErr error
}

type unaryCase struct {
	name    string
	x       float64
	want    float64
	wantErr error
}

func runBinary(t *testing.T, fn func(a, b float64) (float64, error), cases []binaryCase) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := fn(tc.a, tc.b)
			assertResult(t, got, err, tc.want, tc.wantErr)
		})
	}
}

func runUnary(t *testing.T, fn func(x float64) (float64, error), cases []unaryCase) {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := fn(tc.x)
			assertResult(t, got, err, tc.want, tc.wantErr)
		})
	}
}

// assertResult expects wantErr if it is set, and otherwise a result that is
// bit-identical to want, so a -0 result cannot pass as 0.
func assertResult(t *testing.T, got float64, err error, want float64, wantErr error) {
	t.Helper()
	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Fatalf("got (%v, %v), want error %q", got, err, wantErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAdd(t *testing.T) {
	runBinary(t, calculator.Add, []binaryCase{
		{name: "positives", a: 2, b: 3, want: 5},
		{name: "negatives", a: -2, b: -3, want: -5},
		{name: "mixed signs", a: -2, b: 5, want: 3},
		{name: "decimals", a: 1.5, b: 2.25, want: 3.75},
		{name: "decimals keep float64 precision", a: 0.1, b: 0.2, want: 0.30000000000000004},
		{name: "zero", a: 7, b: 0, want: 7},
		{name: "negative zero is normalized", a: negZero, b: negZero, want: 0},
		{name: "overflow", a: math.MaxFloat64, b: math.MaxFloat64, wantErr: calculator.ErrOverflow},
		{name: "negative overflow", a: -math.MaxFloat64, b: -math.MaxFloat64, wantErr: calculator.ErrOverflow},
	})
}

func TestSubtract(t *testing.T) {
	runBinary(t, calculator.Subtract, []binaryCase{
		{name: "positive result", a: 5, b: 3, want: 2},
		{name: "negative result", a: 3, b: 5, want: -2},
		{name: "negatives", a: -5, b: -3, want: -2},
		{name: "decimals", a: 5.5, b: 2.25, want: 3.25},
		{name: "decimals keep float64 precision", a: 0.3, b: 0.1, want: 0.19999999999999998},
		{name: "equal operands", a: 5, b: 5, want: 0},
		{name: "negative zero is normalized", a: negZero, b: 0, want: 0},
		{name: "overflow", a: -math.MaxFloat64, b: math.MaxFloat64, wantErr: calculator.ErrOverflow},
	})
}

func TestMultiply(t *testing.T) {
	runBinary(t, calculator.Multiply, []binaryCase{
		{name: "positives", a: 3, b: 4, want: 12},
		{name: "one negative", a: -3, b: 4, want: -12},
		{name: "two negatives", a: -3, b: -4, want: 12},
		{name: "decimals", a: 2.5, b: 4, want: 10},
		{name: "decimals keep float64 precision", a: 0.1, b: 3, want: 0.30000000000000004},
		{name: "zero", a: 5, b: 0, want: 0},
		{name: "negative zero is normalized", a: -1, b: 0, want: 0},
		{name: "underflow to zero is not an error", a: 1e-200, b: 1e-200, want: 0},
		{name: "overflow", a: math.MaxFloat64, b: 2, wantErr: calculator.ErrOverflow},
		{name: "negative overflow", a: math.MaxFloat64, b: -2, wantErr: calculator.ErrOverflow},
	})
}

func TestDivide(t *testing.T) {
	runBinary(t, calculator.Divide, []binaryCase{
		{name: "exact", a: 10, b: 4, want: 2.5},
		{name: "negative dividend", a: -10, b: 4, want: -2.5},
		{name: "two negatives", a: -10, b: -4, want: 2.5},
		{name: "decimals", a: 7.5, b: 2.5, want: 3},
		{name: "repeating decimal", a: 1, b: 3, want: 1.0 / 3},
		{name: "zero dividend", a: 0, b: 5, want: 0},
		{name: "negative zero is normalized", a: 0, b: -5, want: 0},
		{name: "by zero", a: 10, b: 0, wantErr: calculator.ErrDivisionByZero},
		{name: "by negative zero", a: 10, b: negZero, wantErr: calculator.ErrDivisionByZero},
		{name: "zero by zero", a: 0, b: 0, wantErr: calculator.ErrDivisionByZero},
		{name: "overflow", a: math.MaxFloat64, b: 0.5, wantErr: calculator.ErrOverflow},
	})
}

func TestPower(t *testing.T) {
	runBinary(t, calculator.Power, []binaryCase{
		{name: "positive exponent", a: 2, b: 10, want: 1024},
		{name: "negative exponent", a: 2, b: -2, want: 0.25},
		{name: "fractional exponent", a: 4, b: 0.5, want: 2},
		{name: "decimal base", a: 1.5, b: 2, want: 2.25},
		{name: "negative base, odd exponent", a: -2, b: 3, want: -8},
		{name: "negative base, even exponent", a: -2, b: 2, want: 4},
		{name: "negative base, negative exponent", a: -8, b: -1, want: -0.125},
		{name: "zero exponent", a: 9, b: 0, want: 1},
		{name: "zero to the zero is 1", a: 0, b: 0, want: 1},
		{name: "zero base", a: 0, b: 5, want: 0},
		{name: "zero to a negative power", a: 0, b: -1, wantErr: calculator.ErrDivisionByZero},
		{name: "negative zero to a negative power", a: negZero, b: -1, wantErr: calculator.ErrDivisionByZero},
		{name: "negative base, fractional exponent", a: -8, b: 1.0 / 3, wantErr: calculator.ErrDomain},
		{name: "negative base, square root", a: -2, b: 0.5, wantErr: calculator.ErrDomain},
		{name: "overflow", a: 10, b: 400, wantErr: calculator.ErrOverflow},
		{name: "negative overflow", a: -10, b: 401, wantErr: calculator.ErrOverflow},
	})
}

func TestSqrt(t *testing.T) {
	runUnary(t, calculator.Sqrt, []unaryCase{
		{name: "perfect square", x: 16, want: 4},
		{name: "irrational", x: 2, want: math.Sqrt2},
		{name: "decimal", x: 0.25, want: 0.5},
		{name: "zero", x: 0, want: 0},
		{name: "negative zero is normalized", x: negZero, want: 0},
		{name: "negative", x: -1, wantErr: calculator.ErrDomain},
		{name: "tiny negative", x: -1e-300, wantErr: calculator.ErrDomain},
	})
}

func TestPercentage(t *testing.T) {
	runBinary(t, calculator.Percentage, []binaryCase{
		{name: "15% of 200", a: 15, b: 200, want: 30},
		{name: "over 100%", a: 200, b: 50, want: 100},
		{name: "decimal percent", a: 12.5, b: 80, want: 10},
		{name: "fraction of a percent", a: 0.5, b: 10, want: 0.05},
		{name: "multiplies first to stay exact", a: 7, b: 300, want: 21},
		{name: "negative percent", a: -10, b: 50, want: -5},
		{name: "negative value", a: 10, b: -50, want: -5},
		{name: "zero percent", a: 0, b: 123, want: 0},
		{name: "negative zero is normalized", a: -10, b: 0, want: 0},
		{name: "overflow", a: 1e300, b: 1e300, wantErr: calculator.ErrOverflow},
	})
}
