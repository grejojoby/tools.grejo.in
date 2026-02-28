package handler

import (
	"encoding/json"
	"net/http"
)

// CaesarCipher implements a Vigenère-enhanced Caesar cipher.
// The text key provides per-position shifts (cycling), and the numeric
// shift adds a fixed base offset on top. This degrades to a plain Caesar
// cipher when key is empty (or "a") and only shift is set.
func CaesarCipher(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed")
		return
	}
	var req struct {
		Input string `json:"input"`
		Key   string `json:"key"`   // text key (Vigenère-style)
		Shift int    `json:"shift"` // numeric base shift
		Mode  string `json:"mode"`  // "encode" | "decode"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body")
		return
	}

	// Extract only alphabetic runes from the key (normalise to lowercase).
	var keyRunes []rune
	for _, c := range req.Key {
		switch {
		case c >= 'a' && c <= 'z':
			keyRunes = append(keyRunes, c)
		case c >= 'A' && c <= 'Z':
			keyRunes = append(keyRunes, c+32)
		}
	}
	// Neutral fallback: key 'a' contributes 0 shift.
	if len(keyRunes) == 0 {
		keyRunes = []rune{'a'}
	}

	input := []rune(req.Input)
	result := make([]rune, len(input))
	keyIdx := 0 // advances only for alpha characters

	for i, c := range input {
		isUpper := c >= 'A' && c <= 'Z'
		isLower := c >= 'a' && c <= 'z'
		if !isUpper && !isLower {
			result[i] = c
			continue
		}

		// Combine key-derived shift and numeric shift, normalised to [0, 25].
		keyShift := int(keyRunes[keyIdx%len(keyRunes)] - 'a')
		totalShift := ((keyShift + req.Shift) % 26 + 26) % 26
		if req.Mode == "decode" {
			totalShift = (26 - totalShift) % 26
		}

		base := rune('a')
		if isUpper {
			base = 'A'
		}
		result[i] = base + (c-base+rune(totalShift))%26
		keyIdx++
	}

	writeJSON(w, map[string]string{"output": string(result)})
}
