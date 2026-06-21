// Command dsa-sheet serves the DSA Sheet tracker: a small Go backend with a
// JSON API for problem data and per-session progress, plus the static
// frontend that talks to it. Everything (HTML/CSS/JS + the problem list)
// is embedded into the binary, so deployment is just "run this one binary".
package main

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"dsa-sheet/internal/store"
)

//go:embed web/*
var webFS embed.FS

//go:embed data/problems.json
var problemsRaw []byte

const sessionCookie = "dsa_session"

var problems []Problem

func main() {
	var err error
	problems, err = loadProblems(problemsRaw)
	if err != nil {
		log.Fatalf("failed to load problems: %v", err)
	}

	dataDir := getenv("DATA_DIR", ".")
	st, err := store.New(dataDir + "/progress.json")
	if err != nil {
		log.Fatalf("failed to open progress store: %v", err)
	}

	mux := http.NewServeMux()

	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("failed to open embedded web assets: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	mux.HandleFunc("/api/problems", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, problems)
	})

	mux.HandleFunc("/api/progress", func(w http.ResponseWriter, r *http.Request) {
		sid := ensureSession(w, r)
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, st.Get(sid))
		case http.MethodPost:
			var body struct {
				ProblemID string `json:"problemId"`
				Status    string `json:"status"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if err := st.Set(sid, body.ProblemID, store.Status(body.Status)); err != nil {
				http.Error(w, "could not save", http.StatusInternalServerError)
				return
			}
			writeJSON(w, st.Get(sid))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		sid := ensureSession(w, r)
		if err := st.Reset(sid); err != nil {
			http.Error(w, "could not reset", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]bool{"ok": true})
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	addr := ":" + getenv("PORT", "8080")
	log.Printf("DSA Sheet server listening on %s (problems: %d)", addr, len(problems))
	log.Fatal(http.ListenAndServe(addr, logRequests(mux)))
}

func ensureSession(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie(sessionCookie); err == nil && c.Value != "" {
		return c.Value
	}
	sid := randomID()
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    sid,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return sid
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func logRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
