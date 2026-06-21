package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strings"
)

// Problem is one row of the DSA Sheet.
type Problem struct {
	ID      string `json:"id"`
	Idx     int    `json:"idx"`
	Topic   string `json:"topic"`
	Name    string `json:"name"`
	URL     string `json:"url"`
}

type rawProblem struct {
	Topic   string `json:"topic"`
	Name    string `json:"name"`
	URL     string `json:"url"`
}

func loadProblems(b []byte) ([]Problem, error) {
	var raw []rawProblem
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	out := make([]Problem, 0, len(raw))
	for i, r := range raw {
		out = append(out, Problem{
			ID:      problemID(r.Topic, r.Name),
			Idx:     i + 1,
			Topic:   r.Topic,
			Name:    r.Name,
			URL:     r.URL,
		})
	}
	return out, nil
}

// problemID derives a stable, deterministic ID from topic+name so progress
// stays attached to the right problem even if the list is re-sorted.
func problemID(topic, name string) string {
	h := sha1.New()
	h.Write([]byte(strings.ToLower(topic) + "|" + strings.ToLower(name)))
	return hex.EncodeToString(h.Sum(nil))[:12]
}
