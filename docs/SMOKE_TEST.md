# Smoke test

End-to-end checks of the Docker image, in which the Go server serves both the API and the built
frontend. Run the commands in a POSIX shell: a macOS or Linux terminal, or Git Bash on Windows.

## 1. Build and start

```bash
docker compose up --build -d
```

- [ ] The build succeeds. It runs the frontend tests and `go test -race`; a failing test stops
      the build.
- [ ] `docker compose ps` shows the service `Up`, with port `8080->8080`.
- [ ] `docker compose logs calculator` shows
      `msg=listening addr=[::]:8080 static_dir=/app/public cors_origins=[]`.

## 2. API checks

```bash
bash scripts/smoke-test.sh
```

- [ ] Every line starts with `ok`, the last line is `All checks passed`, and the exit code is 0.

The script checks the health endpoint, the page, every operation and every error case from the
API reference in the [README](../README.md). It waits up to 15 seconds for the server, prints the
expected and actual response for any failure, and exits non-zero if anything fails. CI runs the
same script against the image on every push.

One more check, for CORS: the image sets `CORS_ORIGINS=""`, because the page and the API share
one origin.

```bash
curl -s -o /dev/null -D - -X OPTIONS -H 'Origin: http://localhost:5173' \
  -H 'Access-Control-Request-Method: POST' http://localhost:8080/api/v1/calculate | grep -iE '^HTTP|access-control'
```

- [ ] It prints only `HTTP/1.1 204 No Content`, with no `Access-Control-*` headers.

## 3. Container checks

- [ ] `docker image ls calculator` shows about 16 MB.
- [ ] `docker inspect -f '{{.Config.User}}' calculator:latest` prints `nonroot`.
- [ ] `curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/assets/` prints `404`: the
      static file server doesn't list directories.
- [ ] Graceful shutdown: `docker compose stop` returns within a second or two, the last log
      line is `msg="shutting down"`, and
      `docker inspect -f '{{.State.ExitCode}}' $(docker compose ps -aq calculator)` prints `0`.
      Run `docker compose start` afterwards to continue.

## 4. Manual UI steps

Open <http://localhost:8080>.

- [ ] **Basic:** with **Add**, enter `2` and `3`, then click **Calculate**. The result panel shows
      `2 + 3 =` and `5`.
- [ ] **Display:** `0.1` + `0.2` shows `0.3`, though the API returns `0.30000000000000004`.
      `1000000000000000` + `1` shows `1000000000000001`, with every digit. **Power** `2` and `60`
      shows `1.15292150460685e+18`.
- [ ] **Labels follow the operation:** **Power** relabels the fields **Base** and **Exponent**.
      **Percent** uses **Percentage (%)** and **Of number**: `15` and `200` gives `15% of 200 = 30`.
- [ ] **Unary:** **Square root** hides the second field, and `16` gives `√16 = 4`. Switching back
      to **Add** restores the second field and its value.
- [ ] **API errors:** **Divide** `1` by `0` shows "You can’t divide by zero." **Square root** of
      `-4` shows "You can’t take the square root of a negative number." Editing a field removes
      the message.
- [ ] **Validation:** click **Calculate** with both fields empty. Each field shows
      "Enter a number." with a red border, focus moves to the first field, and the browser's
      network tab shows no request. Then try these inputs:
      - `1.2.3` shows "A number can only have one decimal point."
      - `1,000` and `1,5` show "Use a dot for decimals and no thousands separators, like 1500 or 1.5."
      - `1e-400` shows "This number is too small."
      - `abc` shows "Enter a valid number, like -3.5 or 1e3."
- [ ] **Keyboard only:** reload, then press Tab to focus **Add**, the arrow keys to change the
      operation, Tab to the fields, and Enter to calculate. The focus ring is always visible.
- [ ] **Loading state:** in DevTools → Network, set throttling to a slow profile and calculate.
      The button reads **Calculating…** and the result panel shows "Calculating…". Clicking again
      sends no second request.
- [ ] **Backend down:** with the page open, run `docker compose stop` and calculate. After a
      moment (about 2 seconds on Windows) it shows "Can’t reach the calculator service. Check
      your connection and try again." Run `docker compose start` to continue.
- [ ] **Mobile:** in DevTools' device toolbar at 360px wide, nothing scrolls sideways and the
      operation buttons wrap into rows of four. **Multiply** `-1.7976931348623157e308` by `1`
      shows `-1.7976931348623157e+308` on one line.
- [ ] **Dark mode:** in DevTools → Rendering, emulate `prefers-color-scheme: dark`, or switch the
      OS theme. The colors switch and the text stays readable.

## 5. Clean up

```bash
docker compose down
```
