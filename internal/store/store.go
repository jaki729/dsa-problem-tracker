// Package store persists per-session problem progress (todo / solved / revisit)
// to a JSON file on disk. It's intentionally simple — no external DB — so the
// whole app deploys as a single static binary plus one data file.
package store

import (
	"encoding/json"
	"os"
	"sync"
)

// Status is the state of a single problem for a given session.
type Status string

const (
	Todo    Status = "todo"
	Solved  Status = "solved"
	Revisit Status = "revisit"
)

func Cycle(s Status) Status {
	switch s {
	case Todo:
		return Solved
	case Solved:
		return Revisit
	default:
		return Todo
	}
}

// SessionProgress maps problem ID -> status for one session/browser.
type SessionProgress map[string]Status

// Store holds progress for every session in memory and flushes to disk.
type Store struct {
	mu       sync.RWMutex
	path     string
	sessions map[string]SessionProgress
}

func New(path string) (*Store, error) {
	s := &Store{
		path:     path,
		sessions: map[string]SessionProgress{},
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, &s.sessions)
}

func (s *Store) persist() error {
	b, err := json.MarshalIndent(s.sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}

// Get returns a copy of the progress map for a session.
func (s *Store) Get(sessionID string) SessionProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := SessionProgress{}
	for k, v := range s.sessions[sessionID] {
		out[k] = v
	}
	return out
}

// Set updates one problem's status for a session and persists to disk.
// Setting Todo removes the entry (todo is the default, no need to store it).
func (s *Store) Set(sessionID, problemID string, status Status) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessions[sessionID] == nil {
		s.sessions[sessionID] = SessionProgress{}
	}
	if status == Todo {
		delete(s.sessions[sessionID], problemID)
	} else {
		s.sessions[sessionID][problemID] = status
	}
	return s.persist()
}

// Reset clears all progress for a session.
func (s *Store) Reset(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	return s.persist()
}
