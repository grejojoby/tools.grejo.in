package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
)

func UUIDv4Generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed")
		return
	}

	var req struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body")
		return
	}

	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 100 {
		req.Count = 100
	}

	uuids := make([]string, req.Count)
	for i := range uuids {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			writeError(w, "failed to generate UUID")
			return
		}
		b[6] = (b[6] & 0x0f) | 0x40 // version 4
		b[8] = (b[8] & 0x3f) | 0x80 // variant 10xxxxxx
		uuids[i] = fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	}

	writeJSON(w, map[string]any{"uuids": uuids})
}
