package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/CemGuven1/calculator-assesment/backend/internal/calculator"
)

// maxBodyBytes caps the request body. A valid request is well under 100 bytes.
const maxBodyBytes = 1 << 10

// calculateFields are the only fields a calculate request may contain.
var calculateFields = []string{"operation", "operands"}

type calculateRequest struct {
	Operation string `json:"operation"`
	// Raw, so that each operand can be checked on its own: null, a value
	// that is not a number, or a number that float64 cannot hold.
	Operands []json.RawMessage `json:"operands"`
}

type calculateResponse struct {
	Result float64 `json:"result"`
}

// calculate handles POST /api/v1/calculate.
func (h *handlers) calculate(w http.ResponseWriter, r *http.Request) {
	operation, operands, apiErr := parseCalculateRequest(w, r)
	if apiErr != nil {
		writeError(w, apiErr)
		return
	}
	result, err := calculator.Calculate(operation, operands)
	if err != nil {
		writeError(w, calculatorError(r.Context(), h.logger, err))
		return
	}
	writeJSON(w, http.StatusOK, calculateResponse{Result: result})
}

// parseCalculateRequest checks a request in the order a client would fix it:
// the body as a whole (INVALID_REQUEST), then the operation
// (UNKNOWN_OPERATION), then the operands (INVALID_OPERANDS). Whether the
// number of operands fits the operation is left to the calculator package.
func parseCalculateRequest(w http.ResponseWriter, r *http.Request) (string, []float64, *apiError) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return "", nil, &apiError{http.StatusUnsupportedMediaType, codeUnsupportedMediaType, "Content-Type must be application/json"}
	}
	body, apiErr := readBody(w, r)
	if apiErr != nil {
		return "", nil, apiErr
	}
	if apiErr := checkFields(body, calculateFields); apiErr != nil {
		return "", nil, apiErr
	}
	var req calculateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return "", nil, typeError(err)
	}

	if req.Operation == "" {
		return "", nil, &apiError{http.StatusBadRequest, codeUnknownOperation, "operation is required"}
	}
	if _, err := calculator.Arity(req.Operation); err != nil {
		return "", nil, &apiError{http.StatusBadRequest, codeUnknownOperation, err.Error()}
	}

	operands := make([]float64, len(req.Operands))
	for i, raw := range req.Operands {
		x, apiErr := parseOperand(i, raw)
		if apiErr != nil {
			return "", nil, apiErr
		}
		operands[i] = x
	}
	return req.Operation, operands, nil
}

// readBody reads the whole request body, up to maxBodyBytes.
func readBody(w http.ResponseWriter, r *http.Request) ([]byte, *apiError) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	var tooLarge *http.MaxBytesError
	switch {
	case errors.As(err, &tooLarge):
		return nil, payloadTooLarge(tooLarge.Limit)
	case err != nil:
		return nil, invalidRequest("request body could not be read")
	}
	return body, nil
}

// checkFields checks that body is a single JSON object whose fields are all
// in fields, spelled exactly and given once. encoding/json is not strict
// enough on its own: it matches field names case-insensitively (even with
// DisallowUnknownFields), and a repeated field silently replaces the first.
func checkFields(body []byte, fields []string) *apiError {
	notJSON := invalidRequest("request body is not valid JSON")
	dec := json.NewDecoder(bytes.NewReader(body))

	start, err := dec.Token()
	switch {
	case errors.Is(err, io.EOF):
		return invalidRequest("request body is empty")
	case err != nil:
		return notJSON
	case start != json.Delim('{'):
		return invalidRequest("request body must be a JSON object")
	}

	seen := make(map[string]bool, len(fields))
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return notJSON
		}
		name, _ := key.(string) // inside an object, Token always returns keys as strings
		switch {
		case !slices.Contains(fields, name):
			return invalidRequest(fmt.Sprintf("unknown field %q", name))
		case seen[name]:
			return invalidRequest(fmt.Sprintf("duplicate field %q", name))
		}
		seen[name] = true
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return notJSON
		}
	}
	if _, err := dec.Token(); err != nil { // the closing brace
		return notJSON
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return invalidRequest("request body must contain a single JSON object")
	}
	return nil
}

// typeError describes a field of the wrong type. Once checkFields has passed,
// that is the only way json.Unmarshal can fail, and operation and operands
// are the only fields.
func typeError(err error) *apiError {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) && typeErr.Field == "operation" {
		return &apiError{http.StatusBadRequest, codeUnknownOperation, "operation must be a string"}
	}
	return &apiError{http.StatusBadRequest, codeInvalidOperands, "operands must be an array of numbers"}
}

// parseOperand converts the operand at index i, or explains what is wrong
// with it.
func parseOperand(i int, raw json.RawMessage) (float64, *apiError) {
	invalid := func(problem string) (float64, *apiError) {
		return 0, &apiError{http.StatusBadRequest, codeInvalidOperands, fmt.Sprintf("operands[%d] %s", i, problem)}
	}
	text := string(bytes.TrimSpace(raw))
	switch text[0] {
	case 'n':
		return invalid("must be a number, got null")
	case '"':
		return invalid("must be a number, got a string")
	case 't', 'f':
		return invalid("must be a number, got a boolean")
	case '{':
		return invalid("must be a number, got an object")
	case '[':
		return invalid("must be a number, got an array")
	}

	// What is left is a valid JSON number, so ParseFloat can only fail by
	// overflowing float64.
	x, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return invalid("is too large: " + text)
	}
	// A number too close to zero for float64 parses as 0 without an error.
	mantissa, _, _ := strings.Cut(strings.ToLower(text), "e")
	if x == 0 && strings.ContainsAny(mantissa, "123456789") {
		return invalid("is too small: " + text)
	}
	return x, nil
}
