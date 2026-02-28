package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/* ----------------------------------------------------------------
   JSONFormat — pretty
   ---------------------------------------------------------------- */

func TestJSONFormat_Pretty(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		checkOutput func(t *testing.T, out string)
	}{
		{
			name:  "indents output",
			input: `{"b":2,"a":1}`,
			checkOutput: func(t *testing.T, out string) {
				if !strings.Contains(out, "\n") {
					t.Error("expected newlines in prettified output")
				}
			},
		},
		{
			name:  "preserves key order (b before a)",
			input: `{"b":2,"a":1}`,
			checkOutput: func(t *testing.T, out string) {
				bIdx := strings.Index(out, `"b"`)
				aIdx := strings.Index(out, `"a"`)
				if bIdx == -1 || aIdx == -1 || bIdx >= aIdx {
					t.Errorf("prettify must preserve key order; b should appear before a in:\n%s", out)
				}
			},
		},
		{
			name:  "handles nested object",
			input: `{"outer":{"inner":42}}`,
			checkOutput: func(t *testing.T, out string) {
				if !strings.Contains(out, `"inner"`) {
					t.Error("nested key missing from prettified output")
				}
			},
		},
		{
			name:  "handles array",
			input: `[1,2,3]`,
			checkOutput: func(t *testing.T, out string) {
				if !strings.Contains(out, "1") {
					t.Error("array content missing from prettified output")
				}
			},
		},
		{
			name:    "rejects invalid JSON",
			input:   `{bad json}`,
			wantErr: true,
		},
		{
			name:    "rejects empty input",
			input:   ``,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := postJSON(t, JSONFormat, `{"input":`+jsonStr(tc.input)+`,"mode":"pretty"}`)
			if tc.wantErr {
				assertError(t, res)
				return
			}
			assertNoError(t, res)
			out, _ := res["output"].(string)
			if tc.checkOutput != nil {
				tc.checkOutput(t, out)
			}
		})
	}
}

/* ----------------------------------------------------------------
   JSONFormat — compact
   ---------------------------------------------------------------- */

func TestJSONFormat_Compact(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "strips whitespace from object",
			input: "{\n  \"a\": 1,\n  \"b\": 2\n}",
			want:  `{"a":1,"b":2}`,
		},
		{
			name:  "strips whitespace from array",
			input: "[\n  1,\n  2,\n  3\n]",
			want:  `[1,2,3]`,
		},
		{
			name:  "already compact stays compact",
			input: `{"x":true}`,
			want:  `{"x":true}`,
		},
		{
			name:    "rejects invalid JSON",
			input:   `{bad}`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := postJSON(t, JSONFormat, `{"input":`+jsonStr(tc.input)+`,"mode":"compact"}`)
			if tc.wantErr {
				assertError(t, res)
				return
			}
			assertNoError(t, res)
			assertStr(t, res, "output", tc.want)
		})
	}
}

func TestJSONFormat_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	JSONFormat(w, req)
	var res map[string]any
	decodeInto(t, w, &res)
	assertError(t, res)
}

/* ----------------------------------------------------------------
   JSONValidate
   ---------------------------------------------------------------- */

func TestJSONValidate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
	}{
		{"valid object",             `{"key":"value"}`, true},
		{"valid array",              `[1,2,3]`,         true},
		{"valid null",               `null`,            true},
		{"valid number",             `42`,              true},
		{"valid boolean true",       `true`,            true},
		{"valid boolean false",      `false`,           true},
		{"valid string",             `"hello"`,         true},
		{"valid nested",             `{"a":{"b":[1]}}`, true},
		{"invalid — bare word",      `bad`,             false},
		{"invalid — trailing comma", `{"a":1,}`,        false},
		{"invalid — single quotes",  `{'a':1}`,         false},
		{"invalid — unterminated",   `{"a":`,           false},
		{"invalid — empty",          ``,                false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := postJSON(t, JSONValidate, `{"input":`+jsonStr(tc.input)+`}`)
			valid, _ := res["valid"].(bool)
			if valid != tc.wantValid {
				t.Errorf("valid=%v want=%v (error: %v)", valid, tc.wantValid, res["error"])
			}
			if !tc.wantValid {
				if msg, _ := res["error"].(string); msg == "" {
					t.Error("expected a non-empty error message for invalid JSON")
				}
			}
		})
	}
}

/* ----------------------------------------------------------------
   JSONCompare
   ---------------------------------------------------------------- */

func TestJSONCompare(t *testing.T) {
	tests := []struct {
		name      string
		a, b      string
		wantEqual bool
		// line types that must appear among the returned lines
		wantTypes []string
		wantErr   bool
	}{
		{
			name:      "identical objects",
			a:         `{"name":"Alice","age":30}`,
			b:         `{"name":"Alice","age":30}`,
			wantEqual: true,
		},
		{
			name:      "changed scalar value",
			a:         `{"name":"Alice"}`,
			b:         `{"name":"Bob"}`,
			wantEqual: false,
			wantTypes: []string{"changed", "changed_new"},
		},
		{
			name:      "added key",
			a:         `{"name":"Alice"}`,
			b:         `{"name":"Alice","city":"NYC"}`,
			wantEqual: false,
			wantTypes: []string{"added"},
		},
		{
			name:      "removed key",
			a:         `{"name":"Alice","extra":true}`,
			b:         `{"name":"Alice"}`,
			wantEqual: false,
			wantTypes: []string{"removed"},
		},
		{
			name:      "nested value changed",
			a:         `{"user":{"name":"Alice","age":30}}`,
			b:         `{"user":{"name":"Bob","age":30}}`,
			wantEqual: false,
			wantTypes: []string{"changed", "changed_new"},
		},
		{
			name:      "array element changed",
			a:         `[1,2,3]`,
			b:         `[1,9,3]`,
			wantEqual: false,
			wantTypes: []string{"changed", "changed_new"},
		},
		{
			name:      "array length differs — extra element removed",
			a:         `[1,2,3]`,
			b:         `[1,2]`,
			wantEqual: false,
			wantTypes: []string{"removed"},
		},
		{
			name:      "type changed object→array",
			a:         `{"x":1}`,
			b:         `[1]`,
			wantEqual: false,
			wantTypes: []string{"changed", "changed_new"},
		},
		{
			name:    "invalid JSON in A",
			a:       `{bad}`,
			b:       `{"ok":1}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON in B",
			a:       `{"ok":1}`,
			b:       `{bad}`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := postJSON(t, JSONCompare, `{"a":`+jsonStr(tc.a)+`,"b":`+jsonStr(tc.b)+`}`)
			if tc.wantErr {
				assertError(t, res)
				return
			}
			assertNoError(t, res)

			equal, _ := res["equal"].(bool)
			if equal != tc.wantEqual {
				t.Errorf("equal=%v want=%v; lines=%v", equal, tc.wantEqual, res["lines"])
			}

			lines, _ := res["lines"].([]any)
			typesSeen := map[string]bool{}
			for _, l := range lines {
				lm, _ := l.(map[string]any)
				if typ, _ := lm["type"].(string); typ != "" {
					typesSeen[typ] = true
				}
			}
			for _, wantType := range tc.wantTypes {
				if !typesSeen[wantType] {
					t.Errorf("no line with type %q; types seen: %v", wantType, typesSeen)
				}
			}
		})
	}
}

/* ----------------------------------------------------------------
   JSONSortKeys
   ---------------------------------------------------------------- */

func TestJSONSortKeys(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		checkOutput func(t *testing.T, out string)
	}{
		{
			name:  "sorts top-level keys alphabetically",
			input: `{"z":3,"a":1,"m":2}`,
			checkOutput: func(t *testing.T, out string) {
				aIdx := strings.Index(out, `"a"`)
				mIdx := strings.Index(out, `"m"`)
				zIdx := strings.Index(out, `"z"`)
				if !(aIdx < mIdx && mIdx < zIdx) {
					t.Errorf("keys not in alphabetical order:\n%s", out)
				}
			},
		},
		{
			// json.MarshalIndent emits `"key": value` (space after colon)
			name:  "sorts nested keys at all levels",
			input: `{"b":{"z":9,"a":1},"a":0}`,
			checkOutput: func(t *testing.T, out string) {
				// top-level: "a" key must appear before "b" key
				aTop := strings.Index(out, `"a":`)
				bTop := strings.Index(out, `"b":`)
				if aTop == -1 || bTop == -1 || aTop >= bTop {
					t.Errorf("top-level keys not sorted (a should precede b):\n%s", out)
				}
				// nested: inner "a" must appear before inner "z"
				// Both appear after "b": so search within the substring after bTop
				sub := out[bTop:]
				aInner := strings.Index(sub, `"a":`)
				zInner := strings.Index(sub, `"z":`)
				if aInner == -1 || zInner == -1 || aInner >= zInner {
					t.Errorf("nested keys not sorted (a should precede z):\n%s", out)
				}
			},
		},
		{
			name:  "preserves all array element values",
			input: `[3,1,2]`,
			checkOutput: func(t *testing.T, out string) {
				for _, n := range []string{"1", "2", "3"} {
					if !strings.Contains(out, n) {
						t.Errorf("array element %s missing from output: %s", n, out)
					}
				}
			},
		},
		{
			name:    "rejects invalid JSON",
			input:   `{bad json}`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := postJSON(t, JSONSortKeys, `{"input":`+jsonStr(tc.input)+`}`)
			if tc.wantErr {
				assertError(t, res)
				return
			}
			assertNoError(t, res)
			out, _ := res["output"].(string)
			if tc.checkOutput != nil {
				tc.checkOutput(t, out)
			}
		})
	}
}
