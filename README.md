# DSA Sheet — Go backend + tracker UI

A self-contained progress tracker for your DSA sheet (DSA problems, 18 topics).
Unlike the static version, this one has a real **Go backend**: it serves the
problem list and stores per-session progress (todo / solved / revisit) in a
JSON file on the server, behind a small REST API. The frontend is plain
HTML/CSS/JS that just calls that API.


## Project layout
```
main.go                 — HTTP server, routes, sessions
problems.go             — loads data/problems.json, derives stable problem IDs
internal/store/store.go — JSON-file-backed progress store (thread-safe)
data/problems.json      — your DSA problems (topic, name, url)
web/                    — frontend (index.html, app.js, styles.css), embedded into the binary
Dockerfile              — multi-stage build → small Alpine image
.github/workflows/      — CI + Docker publish
```

## Run with Docker
```bash
docker build -t dsa-sheet .
docker run -p 8080:8080 -v dsa-data:/app/data dsa-sheet
```
The `-v dsa-data:/app/data` volume keeps progress across container restarts.

## API
| Method | Path             | Description                                  |
|--------|------------------|-----------------------------------------------|
| GET    | `/api/problems`  | All DSA problems                              |
| GET    | `/api/progress`  | This session's progress (`{problemId: status}`) |
| POST   | `/api/progress`  | Body `{"problemId":"...","status":"solved"}`  |
| POST   | `/api/reset`     | Clears this session's progress                |
| GET    | `/healthz`       | Liveness check                                |

Sessions are tracked via an `HttpOnly` cookie — no login, but progress is
per-browser unless add real auth later.