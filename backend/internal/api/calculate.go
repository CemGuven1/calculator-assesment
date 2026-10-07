package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/CemGuven1/calculator-assesment/backend/internal/calculator"
)

// maxBodyBytes caps the request body. A valid request is well under 100 bytes.
const maxBodyBytes = 1 << 10

type calculateRequest struct {
	Operation string `json:"operation"`
	// Pointers, so that a null element can be told apart from 0.
	Operands []*float64 `json:"operands"`
}

type calculateResponse struct {
	Result float64 `json:"result"`
}

// calculate handles POST /api/v1/calculate.
func calculate(w http.ResponseWriter, r *http.Request) {
	operation, operands, apiErr := parseCalculateRequest(w, r)
	if apiErr != nil {
		writeError(w, apiErr)
		return
	}
	result, err := calculator.Calculate(operation, operands)
	if err != nil {
		writeError(w, calculatorError(err))
		return
	}
	writeJSON(w, http.StatusOK, calculateResponse{Result: result})
}

// parseCalculateRequest checks the Content-Type, decodes the body strictly and
// checks the shape of the request. Whether the operation exists and gets the
// right number of operands is left to the calculator package.
func parseCalculateRequest(w http.ResponseWriter, r *http.Request) (string, []float64, *apiError) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return "", nil, &apiError{http.StatusUnsupportedMediaType, codeUnsupportedMediaType, "Content-Type must be application/json"}
	}

	var req calculateRequest
	if apiErr := decodeJSON(w, r, &req); apiErr != nil {
		return "", nil, apiErr
	}

	if req.Operation == "" {
		return "", nil, &apiError{http.StatusBadRequest, codeUnknownOperation, "operation is required"}
	}
	operands := make([]float64, len(req.Operands))
	for i, x := range req.Operands {
		if x == nil {
			message := fmt.Sprintf("operands[%d] must be a number, got null", i)
			return "", nil, &apiError{http.StatusBadRequest, codeInvalidOperands, message}
		}
		operands[i] = *x
	}
	return req.Operation, operands, nil
}

// decodeJSON decodes a single JSON object from the request body into dst. It
// rejects unknown fields, trailing data and bodies over maxBodyBytes.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) *apiError {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return decodeError(err)
	}

	// The object may only be followed by whitespace.
	err := dec.Decode(&struct{}{})
	var tooLarge *http.MaxBytesError
	switch {
	case errors.Is(err, io.EOF):
		return nil
	case errors.As(err, &tooLarge):
		return payloadTooLarge(tooLarge.Limit)
	default:
		return invalidRequest("request body must contain a single JSON object")
	}
}

// decodeError turns an error from json.Decoder into a response whose message
// is safe and useful to show to API clients.
func decodeError(err error) *apiError {
	var (
		tooLarge  *http.MaxBytesError
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
	)
	switch {
	case errors.As(err, &tooLarge):
		return payloadTooLarge(tooLarge.Limit)
	case errors.Is(err, io.EOF):
		return invalidRequest("request body is empty")
	case errors.As(err, &syntaxErr), errors.Is(err, io.ErrUnexpectedEOF):
		return invalidRequest("request body is not valid JSON")
	case errors.As(err, &typeErr):
		return invalidRequest(typeErrorMessage(typeErr))
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		// encoding/json has no exported error type for unknown fields.
		return invalidRequest(strings.TrimPrefix(err.Error(), "json: "))
	default:
		return invalidRequest("request body could not be read")
	}
}

// typeErrorMessage describes a JSON value of the wrong type in a
// calculateRequest.
func typeErrorMessage(err *json.UnmarshalTypeError) string {
	switch {
	case strings.HasPrefix(err.Value, "number "):
		// A number too large for float64, such as 1e400.
		return err.Value + " is out of range"
	case err.Field == "operation":
		return "operation must be a string"
	case err.Field == "operands":
		return "operands must be an array of numbers"
	default:
		return "request body must be a JSON object"
	}
}
