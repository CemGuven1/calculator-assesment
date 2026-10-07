package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/CemGuven1/calculator-assesment/backend/internal/calculator"
)

// Error codes sent in the "code" field of every error response. Clients can
// rely on them; the messages are for people and may change.
const (
	codeInvalidRequest       = "INVALID_REQUEST"
	codeUnknownOperation     = "UNKNOWN_OPERATION"
	codeInvalidOperands      = "INVALID_OPERANDS"
	codeDivisionByZero       = "DIVISION_BY_ZERO"
	codeDomainError          = "DOMAIN_ERROR"
	codeOverflow             = "OVERFLOW"
	codePayloadTooLarge      = "PAYLOAD_TOO_LARGE"
	codeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	codeMethodNotAllowed     = "METHOD_NOT_ALLOWED"
	codeNotFound             = "NOT_FOUND"
	codeInternal             = "INTERNAL"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// apiError is a failed request: the HTTP status, and the code and message for
// the JSON error body.
type apiError struct {
	status  int
	code    string
	message string
}

func invalidRequest(message string) *apiError {
	return &apiError{http.StatusBadRequest, codeInvalidRequest, message}
}

func payloadTooLarge(limit int64) *apiError {
	message := fmt.Sprintf("request body must not exceed %d bytes", limit)
	return &apiError{http.StatusRequestEntityTooLarge, codePayloadTooLarge, message}
}

func internalError() *apiError {
	return &apiError{http.StatusInternalServerError, codeInternal, "internal server error"}
}

// calculatorErrors maps errors from the calculator package to responses.
// 400 means the request is invalid; 422 means it is valid but the math has no
// real, finite answer.
var calculatorErrors = []struct {
	err    error
	status int
	code   string
}{
	{calculator.ErrUnknownOperation, http.StatusBadRequest, codeUnknownOperation},
	{calculator.ErrInvalidOperand, http.StatusBadRequest, codeInvalidOperands},
	{calculator.ErrDivisionByZero, http.StatusUnprocessableEntity, codeDivisionByZero},
	{calculator.ErrDomain, http.StatusUnprocessableEntity, codeDomainError},
	{calculator.ErrOverflow, http.StatusUnprocessableEntity, codeOverflow},
}

func calculatorError(err error) *apiError {
	for _, m := range calculatorErrors {
		if errors.Is(err, m.err) {
			return &apiError{m.status, m.code, err.Error()}
		}
	}
	// Only reachable if the calculator gains an error with no mapping above.
	return internalError()
}

// writeJSON sends body as JSON with the given status.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	// Encoding these response types cannot fail. A write error means the
	// client has gone away, so there is no one left to report it to.
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, e *apiError) {
	writeJSON(w, e.status, errorResponse{Error: errorBody{Code: e.code, Message: e.message}})
}
