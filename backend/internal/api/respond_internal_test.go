package api

import (
	"errors"
	"net/http"
	"testing"
)

// An error the mapping doesn't know about must become a 500 that hides its
// details from the client.
func TestUnmappedCalculatorErrorIsInternal(t *testing.T) {
	got := calculatorError(errors.New("some new failure"))

	want := apiError{http.StatusInternalServerError, codeInternal, "internal server error"}
	if *got != want {
		t.Errorf("got %+v, want %+v", *got, want)
	}
}
