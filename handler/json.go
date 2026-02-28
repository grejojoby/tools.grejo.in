package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

// JSONFormat handles prettify (pretty) and compress (compact) modes.
// Uses json.Indent / json.Compact to preserve original key order.
func JSONFormat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed")
		return
	}
	var req struct {
		Input string `json:"input"`
		Mode  string `json:"mode"` // "pretty" | "compact"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body")
		return
	}
	if !json.Valid([]byte(req.Input)) {
		writeError(w, "invalid JSON — check for syntax errors")
		return
	}
	var buf bytes.Buffer
	var err error
	if req.Mode == "compact" {
		err = json.Compact(&buf, []byte(req.Input))
	} else {
		err = json.Indent(&buf, []byte(req.Input), "", "  ")
	}
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeJSON(w, map[string]string{"output": buf.String()})
}

// JSONValidate checks whether input is valid JSON.
func JSONValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed")
		return
	}
	var req struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body")
		return
	}
	var v any
	if err := json.Unmarshal([]byte(req.Input), &v); err != nil {
		writeJSON(w, map[string]any{"valid": false, "error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"valid": true, "error": ""})
}

/* ── Compare ─────────────────────────────────────────────────────────────── */

// diffLine is one annotated line in the inline diff view.
type diffLine struct {
	Text string `json:"text"`
	// "unchanged" | "added" | "removed" | "changed" | "changed_new"
	Type string `json:"type"`
}

// JSONCompare sorts the keys of both documents, then produces an annotated
// line-by-line diff where every line carries a semantic type.
func JSONCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed")
		return
	}
	var req struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body")
		return
	}
	var a, b any
	if err := json.Unmarshal([]byte(req.A), &a); err != nil {
		writeError(w, fmt.Sprintf("invalid JSON A: %s", err.Error()))
		return
	}
	if err := json.Unmarshal([]byte(req.B), &b); err != nil {
		writeError(w, fmt.Sprintf("invalid JSON B: %s", err.Error()))
		return
	}

	lines := diffAny(a, b, "", "", true)
	equal := true
	for _, l := range lines {
		if l.Type != "unchanged" {
			equal = false
			break
		}
	}
	writeJSON(w, map[string]any{"equal": equal, "lines": lines})
}

// diffAny recursively compares two JSON values and returns annotated lines.
// key is the JSON key for this value (empty for array elements / root).
// indent is the current indentation string.
// last controls whether to append a trailing comma.
func diffAny(a, b any, key, indent string, last bool) []diffLine {
	if reflect.DeepEqual(a, b) {
		return emitValue(key, a, "unchanged", indent, last)
	}

	aMap, aIsMap := a.(map[string]any)
	bMap, bIsMap := b.(map[string]any)
	if aIsMap && bIsMap {
		return diffMaps(key, aMap, bMap, indent, last)
	}

	aSlice, aIsSlice := a.([]any)
	bSlice, bIsSlice := b.([]any)
	if aIsSlice && bIsSlice {
		return diffSlices(key, aSlice, bSlice, indent, last)
	}

	// Different types or different scalar values: old line → new line.
	out := emitValue(key, a, "changed", indent, last)
	out = append(out, emitValue(key, b, "changed_new", indent, last)...)
	return out
}

// diffMaps walks both maps in sorted-key order and annotates each entry.
func diffMaps(key string, a, b map[string]any, indent string, last bool) []diffLine {
	childIndent := indent + "  "
	allKeys := mergedSortedKeys(a, b)

	var out []diffLine

	// Opening brace
	opening := indent
	if key != "" {
		kb, _ := json.Marshal(key)
		opening += string(kb) + ": "
	}
	out = append(out, diffLine{Text: opening + "{", Type: "unchanged"})

	for i, k := range allKeys {
		childLast := i == len(allKeys)-1
		aVal, aOk := a[k]
		bVal, bOk := b[k]
		switch {
		case aOk && bOk:
			out = append(out, diffAny(aVal, bVal, k, childIndent, childLast)...)
		case aOk:
			out = append(out, emitValue(k, aVal, "removed", childIndent, childLast)...)
		default:
			out = append(out, emitValue(k, bVal, "added", childIndent, childLast)...)
		}
	}

	// Closing brace
	closing := "}"
	if !last {
		closing += ","
	}
	out = append(out, diffLine{Text: indent + closing, Type: "unchanged"})
	return out
}

// diffSlices compares arrays element-by-element at each position.
func diffSlices(key string, a, b []any, indent string, last bool) []diffLine {
	childIndent := indent + "  "

	var out []diffLine

	// Opening bracket
	opening := indent
	if key != "" {
		kb, _ := json.Marshal(key)
		opening += string(kb) + ": "
	}
	out = append(out, diffLine{Text: opening + "[", Type: "unchanged"})

	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	for i := 0; i < maxLen; i++ {
		childLast := i == maxLen-1
		switch {
		case i < len(a) && i < len(b):
			out = append(out, diffAny(a[i], b[i], "", childIndent, childLast)...)
		case i < len(a):
			out = append(out, emitValue("", a[i], "removed", childIndent, childLast)...)
		default:
			out = append(out, emitValue("", b[i], "added", childIndent, childLast)...)
		}
	}

	// Closing bracket
	closing := "]"
	if !last {
		closing += ","
	}
	out = append(out, diffLine{Text: indent + closing, Type: "unchanged"})
	return out
}

// emitValue serializes val as pretty-printed lines, all tagged with lineType.
// The first line is prefixed with indent+key; continuation lines are indented.
func emitValue(key string, val any, lineType, indent string, last bool) []diffLine {
	b, _ := json.MarshalIndent(val, "", "  ")
	rawLines := strings.Split(string(b), "\n")

	keyPart := ""
	if key != "" {
		kb, _ := json.Marshal(key)
		keyPart = string(kb) + ": "
	}

	out := make([]diffLine, len(rawLines))
	for i, line := range rawLines {
		var text string
		if i == 0 {
			text = indent + keyPart + line
		} else {
			text = indent + line
		}
		if i == len(rawLines)-1 && !last {
			text += ","
		}
		out[i] = diffLine{Text: text, Type: lineType}
	}
	return out
}

// mergedSortedKeys returns the union of keys from both maps, sorted.
func mergedSortedKeys(a, b map[string]any) []string {
	seen := make(map[string]bool, len(a)+len(b))
	for k := range a {
		seen[k] = true
	}
	for k := range b {
		seen[k] = true
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

/* ── Sort Keys ───────────────────────────────────────────────────────────── */

// JSONSortKeys parses then re-marshals JSON — Go's encoder sorts map keys alphabetically at all levels.
func JSONSortKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed")
		return
	}
	var req struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body")
		return
	}
	var v any
	if err := json.Unmarshal([]byte(req.Input), &v); err != nil {
		writeError(w, fmt.Sprintf("invalid JSON: %s", err.Error()))
		return
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeJSON(w, map[string]string{"output": string(out)})
}
