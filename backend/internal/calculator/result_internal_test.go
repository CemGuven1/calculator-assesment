package calculator

import (
	"errors"
	"math"
	"testing"
)

// No operation reaches the NaN branch today, because explicit checks reject
// every NaN-producing input first. The guard protects operations added later.
func TestCheckResultRejectsNaN(t *testing.T) {
	if _, err := checkResult(math.NaN()); !errors.Is(err, ErrDomain) {
		t.Fatalf("err = %v, want %v", err, ErrDomain)
	}
}
