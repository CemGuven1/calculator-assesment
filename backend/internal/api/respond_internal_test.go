package api

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

// An error the mapping doesn't know about must become a 500 that hides its
// details from the client, and those details must reach the log.
func TestUnmappedCalculatorErrorIsInternalAndLogged(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	got := calculatorError(context.Background(), logger, errors.New("some new failure"))

	want := apiError{http.StatusInternalServerError, codeInternal, "internal server error"}
	if *got != want {
		t.Errorf("got %+v, want %+v", *got, want)
	}
	if !strings.Contains(logs.String(), `level=ERROR msg="calculator error with no API mapping" err="some new failure"`) {
		t.Errorf("log = %q, want the error and its cause", logs.String())
	}
}
