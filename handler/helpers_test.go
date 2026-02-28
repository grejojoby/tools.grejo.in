package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func decodeInto(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("could not decode response body: %v", err)
	}
}

// postJSON fires a POST request through a handler and returns the decoded JSON response.
func postJSON(t *testing.T, h http.HandlerFunc, body string) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h(w, req)
	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v\nbody: %s", err, w.Body.String())
	}
	return result
}

// jsonBody encodes a value as a JSON string literal (for embedding in a larger JSON body).
func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// assertStr checks that a map field matches an expected string value.
func assertStr(t *testing.T, res map[string]any, field, want string) {
	t.Helper()
	got, _ := res[field].(string)
	if got != want {
		t.Errorf("field %q: got %q, want %q", field, got, want)
	}
}

// assertNoError fails if the response contains a non-empty "error" field.
func assertNoError(t *testing.T, res map[string]any) {
	t.Helper()
	if err, _ := res["error"].(string); err != "" {
		t.Fatalf("unexpected error: %s", err)
	}
}

// assertError fails if the response does NOT contain a non-empty "error" field.
func assertError(t *testing.T, res map[string]any) {
	t.Helper()
	if res["error"] == nil || res["error"] == "" {
		t.Errorf("expected non-empty error field, got: %v", res)
	}
}
