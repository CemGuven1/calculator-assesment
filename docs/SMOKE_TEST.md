# Smoke test

End-to-end checks of the Docker image: the Go server serving both the API and the built
frontend. The expected outputs below were recorded from a real run.

Run the `curl` commands in a POSIX shell: a macOS or Linux terminal, or Git Bash on Windows
(in PowerShell, `curl` is an alias for a different command).

## 1. Build and start

```bash
docker compose up --build -d
```

- [ ] The build succeeds. It runs the frontend tests and `go test -race`; a failing test
      stops the build.
- [ ] `docker compose ps` shows the service `Up`, with port `8080->8080`.
- [ ] `docker compose logs calculator` shows
      `msg=listening addr=[::]:8080 static_dir=/app/public cors_origins=[]`.

## 2. API checks with curl

Set up two helpers first:

```bash
BASE=http://localhost:8080
calc() { curl -s -w ' %{http_code}\n' -H 'Content-Type: application/json' -d "$1" "$BASE/api/v1/calculate"; }
```

### Server and static files

| Command | Expected |
|---|---|
| `curl -s -w ' %{http_code}\n' $BASE/health` | `{"status":"ok"} 200` |
| `curl -s $BASE/ \| grep -o '<title>.*</title>'` | `<title>Calculator</title>` |
| `curl -s -o /dev/null -w '%{http_code} %{content_type}\n' $BASE$(curl -s $BASE/ \| grep -o '/assets/[^"]*\.js')` | `200 text/javascript; charset=utf-8` |

### Every operation

| Command | Expected |
|---|---|
| `calc '{"operation":"add","operands":[2,3]}'` | `{"result":5} 200` |
| `calc '{"operation":"subtract","operands":[5,8]}'` | `{"result":-3} 200` |
| `calc '{"operation":"multiply","operands":[4,2.5]}'` | `{"result":10} 200` |
| `calc '{"operation":"divide","operands":[10,4]}'` | `{"result":2.5} 200` |
| `calc '{"operation":"power","operands":[2,10]}'` | `{"result":1024} 200` |
| `calc '{"operation":"sqrt","operands":[16]}'` | `{"result":4} 200` |
| `calc '{"operation":"percentage","operands":[15,200]}'` | `{"result":30} 200` |
| `calc '{"operation":"add","operands":[0.1,0.2]}'` | `{"result":0.30000000000000004} 200` (the UI shows `0.3`) |

### Math errors: 422

| Command | Expected |
|---|---|
| `calc '{"operation":"divide","operands":[1,0]}'` | `DIVISION_BY_ZERO`, `division by zero` |
| `calc '{"operation":"sqrt","operands":[-4]}'` | `DOMAIN_ERROR`, `result is not a real number: square root of a negative number` |
| `calc '{"operation":"multiply","operands":[1e308,10]}'` | `OVERFLOW`, `result is out of range` |

### Invalid requests

| Command | Expected |
|---|---|
| `calc '{"operation":"modulo","operands":[1,2]}'` | 400 `UNKNOWN_OPERATION`, lists the supported operations |
| `calc '{"operation":"sqrt","operands":[4,9]}'` | 400 `INVALID_OPERANDS`, `invalid operand: sqrt expects 1 operand, got 2` |
| `calc '{"operation":"add","operands":[1,null]}'` | 400 `INVALID_OPERANDS`, `operands[1] must be a number, got null` |
| `calc '{"operation":"add","operands":[1,2],"extra":1}'` | 400 `INVALID_REQUEST`, `unknown field "extra"` |
| `calc '{"operation":'` | 400 `INVALID_REQUEST`, `request body is not valid JSON` |
| `curl -s -w ' %{http_code}\n' -d '{"operation":"add","operands":[1,2]}' $BASE/api/v1/calculate` | 415 `UNSUPPORTED_MEDIA_TYPE` (no `Content-Type` header) |
| `calc "{\"operation\":\"add\",$(printf '%*s' 2000 '')\"operands\":[1,2]}"` | 413 `PAYLOAD_TOO_LARGE`, `request body must not exceed 1024 bytes` |
| `curl -s -i $BASE/api/v1/calculate \| grep -iE '^(HTTP\|allow)'` | `405 Method Not Allowed` and `Allow: POST` |
| `curl -s -w ' %{http_code}\n' $BASE/api/v1/nope` | `{"error":{"code":"NOT_FOUND","message":"not found"}} 404` |

### CORS is off in the container

The page and the API share one origin, so the image sets `CORS_ORIGINS=""`.

```bash
curl -s -o /dev/null -D - -X OPTIONS -H 'Origin: http://localhost:5173' \
  -H 'Access-Control-Request-Method: POST' $BASE/api/v1/calculate | grep -iE '^HTTP|access-control'
```

- [ ] Prints only `HTTP/1.1 204 No Content`, with no `Access-Control-*` headers.

## 3. Container checks

- [ ] `docker image ls calculator` shows about 16 MB.
- [ ] `docker inspect -f '{{.Config.User}}' calculator:latest` prints `nonroot`.
- [ ] Graceful shutdown: `docker compose stop` returns within a second or two, the last log
      line is `msg="shutting down"`, and
      `docker inspect -f '{{.State.ExitCode}}' $(docker compose ps -aq calculator)` prints `0`.
      Run `docker compose start` afterwards to continue.

## 4. Manual UI steps

Open <http://localhost:8080>.

- [ ] **Basic:** with **Add**, enter `2` and `3`, then click **Calculate**. The result panel shows
      `2 + 3 =` and `5`.
- [ ] **Display rounding:** `0.1` + `0.2` shows `0.3`. The API returned `0.30000000000000004`.
- [ ] **Labels follow the operation:** **Power** relabels the fields **Base** and **Exponent**.
      `2` and `10` gives `2 ^ 10 = 1024`. **Percent** uses **Percentage (%)** and **Of number**.
      `15` and `200` gives `15% of 200 = 30`.
- [ ] **Unary:** **Square root** hides the second field. `16` gives `√16 = 4`.
      Switching back to **Add** restores the second field and its value.
- [ ] **API errors:** **Divide** `1` by `0` shows "You can’t divide by zero." **Square root** of
      `-4` shows "You can’t take the square root of a negative number." Editing a field
      removes the message.
- [ ] **Validation:** click **Calculate** with both fields empty. Each field shows
      "Enter a number." with a red border, focus moves to the first field, and the network tab
      shows no request. `1.2.3` shows "A number can only have one decimal point." `1,5` shows
      "Use a dot for decimals, like 1.5." `abc` shows "Enter a valid number, like -3.5 or 1e3."
- [ ] **Keyboard only:** reload, then press Tab to focus **Add**, the arrow keys to change the
      operation, Tab to the fields, and Enter to calculate. The focus ring is always visible.
- [ ] **Loading state:** in DevTools → Network, set throttling to "Slow 4G" or "3G" and
      calculate. The button reads **Calculating…** and the result panel shows "Calculating…".
      Clicking again sends no second request.
- [ ] **Backend down:** with the page open, run `docker compose stop` and calculate. After a
      moment (about 2 seconds on Windows) it shows "Can’t reach the calculator service. Check
      your connection and try again." Run `docker compose start` to continue.
- [ ] **Mobile:** in DevTools' device toolbar at 360px wide, nothing scrolls sideways and the
      operation buttons wrap into rows of four. **Multiply** `-1.7976931348623157e308` by `1`
      shows `-1.7976931348623157e+308` on one line.
- [ ] **Dark mode:** in DevTools → Rendering, emulate `prefers-color-scheme: dark` (or switch
      the OS theme). The colors switch and the text stays readable.

## 5. Clean up

```bash
docker compose down
```
