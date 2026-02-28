package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

/* ----------------------------------------------------------------
   URLParse
   ---------------------------------------------------------------- */

func TestURLParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, res map[string]any)
	}{
		{
			name:  "full URL with all components",
			input: "https://user:pass@example.com:8080/api/v1?foo=bar&baz=qux#section",
			check: func(t *testing.T, res map[string]any) {
				assertStr(t, res, "scheme",   "https")
				assertStr(t, res, "host",     "example.com")
				assertStr(t, res, "port",     "8080")
				assertStr(t, res, "path",     "/api/v1")
				assertStr(t, res, "fragment", "section")
				assertStr(t, res, "username", "user")
				assertStr(t, res, "password", "pass")
				query, _ := res["query"].([]any)
				if len(query) != 2 {
					t.Errorf("expected 2 query params, got %d: %v", len(query), query)
				}
			},
		},
		{
			name:  "simple http URL",
			input: "http://example.com",
			check: func(t *testing.T, res map[string]any) {
				assertStr(t, res, "scheme", "http")
				assertStr(t, res, "host",   "example.com")
				assertStr(t, res, "port",   "")
				assertStr(t, res, "path",   "")
			},
		},
		{
			name:  "URL with multiple query values for same key",
			input: "https://example.com?tag=go&tag=web&tag=tools",
			check: func(t *testing.T, res map[string]any) {
				query, _ := res["query"].([]any)
				if len(query) != 3 {
					t.Errorf("expected 3 query entries (multi-value key), got %d", len(query))
				}
			},
		},
		{
			name:  "URL without port",
			input: "https://example.com/path",
			check: func(t *testing.T, res map[string]any) {
				assertStr(t, res, "port", "")
				assertStr(t, res, "path", "/path")
			},
		},
		{
			name:  "URL without auth returns empty username and password",
			input: "https://example.com/",
			check: func(t *testing.T, res map[string]any) {
				assertStr(t, res, "username", "")
				assertStr(t, res, "password", "")
			},
		},
		{
			name:  "raw query is preserved",
			input: "https://example.com?a=1&b=2",
			check: func(t *testing.T, res map[string]any) {
				rawQuery, _ := res["raw_query"].(string)
				if rawQuery == "" {
					t.Error("raw_query should not be empty")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := postJSON(t, URLParse, `{"input":`+jsonStr(tc.input)+`}`)
			if tc.wantErr {
				assertError(t, res)
				return
			}
			assertNoError(t, res)
			if tc.check != nil {
				tc.check(t, res)
			}
		})
	}
}

func TestURLParse_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	URLParse(w, req)
	var res map[string]any
	decodeInto(t, w, &res)
	assertError(t, res)
}

/* ----------------------------------------------------------------
   URLEncode
   ---------------------------------------------------------------- */

func TestURLEncode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"space becomes plus",     "hello world",  "hello+world"},
		{"slash encoded",          "a/b",          "a%2Fb"},
		{"equals encoded",         "foo=bar",      "foo%3Dbar"},
		{"ampersand encoded",      "a&b",          "a%26b"},
		{"colon encoded",          "key:value",    "key%3Avalue"},
		{"hash encoded",           "a#b",          "a%23b"},
		{"empty string",           "",             ""},
		{"already safe chars",     "hello123",     "hello123"},
		{"unicode",                "héllo",        "h%C3%A9llo"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := postJSON(t, URLEncode, `{"input":`+jsonStr(tc.input)+`}`)
			assertNoError(t, res)
			assertStr(t, res, "output", tc.want)
		})
	}
}

/* ----------------------------------------------------------------
   URLDecode
   ---------------------------------------------------------------- */

func TestURLDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"plus to space",       "hello+world",   "hello world",  false},
		{"percent slash",       "a%2Fb",         "a/b",          false},
		{"percent equals",      "foo%3Dbar",     "foo=bar",      false},
		{"percent ampersand",   "a%26b",         "a&b",          false},
		{"percent colon",       "key%3Avalue",   "key:value",    false},
		{"unicode sequence",    "h%C3%A9llo",    "héllo",        false},
		{"empty string",        "",              "",             false},
		{"invalid hex digits",  "bad%ZZseq",     "",             true},
		{"truncated sequence",  "bad%2",         "",             true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := postJSON(t, URLDecode, `{"input":`+jsonStr(tc.input)+`}`)
			if tc.wantErr {
				assertError(t, res)
				return
			}
			assertNoError(t, res)
			assertStr(t, res, "output", tc.want)
		})
	}
}

/* ----------------------------------------------------------------
   Encode → Decode roundtrip
   ---------------------------------------------------------------- */

func TestURLEncodeDecodeRoundtrip(t *testing.T) {
	inputs := []string{
		"hello world",
		"foo=bar&baz=qux",
		"special: !@#$%^*()",
		"path/to/resource",
		"key:value pairs",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			encRes := postJSON(t, URLEncode, `{"input":`+jsonStr(input)+`}`)
			assertNoError(t, encRes)
			encoded, _ := encRes["output"].(string)

			decRes := postJSON(t, URLDecode, `{"input":`+jsonStr(encoded)+`}`)
			assertNoError(t, decRes)
			decoded, _ := decRes["output"].(string)

			if decoded != input {
				t.Errorf("roundtrip(%q):\n  encoded=%q\n  decoded=%q", input, encoded, decoded)
			}
		})
	}
}
