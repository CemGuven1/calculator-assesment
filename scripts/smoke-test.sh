#!/usr/bin/env bash
# Smoke-tests a running calculator over HTTP: the health check, the page, and
# every operation and error case listed in the README. Exits non-zero if any
# check fails.
#
#   bash scripts/smoke-test.sh                          # http://localhost:8080
#   bash scripts/smoke-test.sh http://localhost:9090
set -euo pipefail

base=${1:-http://localhost:8080}
failures=0

# Wait up to 15 seconds for the server to start.
deadline=$((SECONDS + 15))
until curl -fs "$base/health" >/dev/null; do
  if ((SECONDS >= deadline)); then
    echo "FAIL  no answer from $base/health"
    exit 1
  fi
  sleep 0.5
done

# check NAME EXPECTED CURL_ARGS...
# EXPECTED is the response body followed by the status code, for example
# '{"result":5} 200'. The body's trailing newline is removed before comparing.
check() {
  local name=$1 want=$2 got
  shift 2
  got=$(curl -s -w ' %{http_code}' "$@" | tr -d '\n') || true
  if [[ $got == "$want" ]]; then
    echo "ok    $name"
  else
    printf 'FAIL  %s\n      want: %s\n      got:  %s\n' "$name" "$want" "$got"
    failures=$((failures + 1))
  fi
}

# calc NAME REQUEST_BODY EXPECTED: POST a JSON body to the calculate endpoint.
calc() {
  check "$1" "$3" -H 'Content-Type: application/json' -d "$2" "$base/api/v1/calculate"
}

err() { # err CODE MESSAGE: the JSON error body
  printf '{"error":{"code":"%s","message":"%s"}}' "$1" "$2"
}

echo "== server"
check "GET /health" '{"status":"ok"} 200' "$base/health"
if curl -fs "$base/" | grep -q '<title>Calculator</title>'; then
  echo "ok    GET / serves the calculator page"
else
  echo "skip  GET / (this server does not serve the frontend; set STATIC_DIR)"
fi

echo "== operations"
calc "add"        '{"operation":"add","operands":[2,3]}'          '{"result":5} 200'
calc "subtract"   '{"operation":"subtract","operands":[5,8]}'     '{"result":-3} 200'
calc "multiply"   '{"operation":"multiply","operands":[4,2.5]}'   '{"result":10} 200'
calc "divide"     '{"operation":"divide","operands":[10,4]}'      '{"result":2.5} 200'
calc "power"      '{"operation":"power","operands":[2,10]}'       '{"result":1024} 200'
calc "sqrt"       '{"operation":"sqrt","operands":[16]}'          '{"result":4} 200'
calc "percentage" '{"operation":"percentage","operands":[15,200]}' '{"result":30} 200'

echo "== math errors (422)"
calc "divide by zero" '{"operation":"divide","operands":[1,0]}' \
  "$(err DIVISION_BY_ZERO 'division by zero') 422"
calc "zero to a negative power" '{"operation":"power","operands":[0,-1]}' \
  "$(err DIVISION_BY_ZERO 'division by zero: zero raised to a negative power') 422"
calc "square root of a negative" '{"operation":"sqrt","operands":[-4]}' \
  "$(err DOMAIN_ERROR 'result is not a real number: square root of a negative number') 422"
calc "negative base, fractional exponent" '{"operation":"power","operands":[-8,0.5]}' \
  "$(err DOMAIN_ERROR 'result is not a real number: negative base with a fractional exponent') 422"
calc "overflow" '{"operation":"multiply","operands":[1e308,10]}' \
  "$(err OVERFLOW 'result is out of range') 422"

echo "== invalid requests (400)"
calc "unknown operation" '{"operation":"modulo","operands":[1,2]}' \
  "$(err UNKNOWN_OPERATION 'unknown operation \"modulo\"; supported operations: add, divide, multiply, percentage, power, sqrt, subtract') 400"
calc "missing operation" '{"operands":[1,2]}' \
  "$(err UNKNOWN_OPERATION 'operation is required') 400"
calc "wrong number of operands" '{"operation":"sqrt","operands":[4,9]}' \
  "$(err INVALID_OPERANDS 'invalid operand: sqrt expects 1 operand, got 2') 400"
calc "null operand" '{"operation":"add","operands":[1,null]}' \
  "$(err INVALID_OPERANDS 'operands[1] must be a number, got null') 400"
calc "operand is a string" '{"operation":"add","operands":["2",3]}' \
  "$(err INVALID_OPERANDS 'operands[0] must be a number, got a string') 400"
calc "operand too large" '{"operation":"add","operands":[1e400,1]}' \
  "$(err INVALID_OPERANDS 'operands[0] is too large: 1e400') 400"
calc "operand too small" '{"operation":"divide","operands":[1,1e-400]}' \
  "$(err INVALID_OPERANDS 'operands[1] is too small: 1e-400') 400"
calc "malformed JSON" '{"operation":' \
  "$(err INVALID_REQUEST 'request body is not valid JSON') 400"
calc "not an object" '[2,3]' \
  "$(err INVALID_REQUEST 'request body must be a JSON object') 400"
calc "unknown field" '{"operation":"add","operands":[1,2],"precision":2}' \
  "$(err INVALID_REQUEST 'unknown field \"precision\"') 400"
calc "field name in the wrong case" '{"Operation":"add","operands":[1,2]}' \
  "$(err INVALID_REQUEST 'unknown field \"Operation\"') 400"
calc "duplicate field" '{"operation":"add","operands":[1,2],"operands":[3,4]}' \
  "$(err INVALID_REQUEST 'duplicate field \"operands\"') 400"

echo "== other HTTP errors"
check "body over 1 KiB (413)" \
  "$(err PAYLOAD_TOO_LARGE 'request body must not exceed 1024 bytes') 413" \
  -H 'Content-Type: application/json' \
  -d "{\"operation\":\"add\",$(printf '%*s' 2000 '')\"operands\":[1,2]}" "$base/api/v1/calculate"
check "no JSON Content-Type (415)" \
  "$(err UNSUPPORTED_MEDIA_TYPE 'Content-Type must be application/json') 415" \
  -d '{"operation":"add","operands":[1,2]}' "$base/api/v1/calculate"
check "wrong method (405)" \
  "$(err METHOD_NOT_ALLOWED 'method GET is not allowed; use POST') 405" \
  "$base/api/v1/calculate"
check "unknown API path (404)" \
  "$(err NOT_FOUND 'not found') 404" \
  "$base/api/v1/nope"

echo
if ((failures > 0)); then
  echo "$failures check(s) failed"
  exit 1
fi
echo "All checks passed"
