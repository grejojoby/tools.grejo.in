package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

/* ----------------------------------------------------------------
   CaesarCipher — pure numeric shift (no text key)
   ---------------------------------------------------------------- */

func TestCaesarCipher_PureShift(t *testing.T) {
	tests := []struct {
		name  string
		input string
		shift int
		mode  string
		want  string
	}{
		// ROT13
		{"ROT13 encode Hello",     "Hello",  13, "encode", "Uryyb"},
		{"ROT13 encode lowercase", "hello",  13, "encode", "uryyb"},
		{"ROT13 decode Uryyb",     "Uryyb",  13, "decode", "Hello"},

		// Classic shift-3
		{"shift-3 encode Hello",   "Hello",  3,  "encode", "Khoor"},
		{"shift-3 decode Khoor",   "Khoor",  3,  "decode", "Hello"},

		// Edge shifts
		{"shift-0 is identity",    "Hello",  0,  "encode", "Hello"},
		{"shift-26 is identity",   "Hello",  26, "encode", "Hello"},
		{"shift-52 is identity",   "Hello",  52, "encode", "Hello"},

		// Wrap-around
		{"wrap around z→a",        "xyz",    3,  "encode", "abc"},
		{"wrap around Z→A",        "XYZ",    3,  "encode", "ABC"},

		// Non-alpha characters are preserved
		{"preserves spaces",       "Hello World",  13, "encode", "Uryyb Jbeyq"},
		{"preserves punctuation",  "Hello, World!", 13, "encode", "Uryyb, Jbeyq!"},
		{"preserves digits",       "H3llo W0rld",   13, "encode", "U3yyb J0eyq"},

		// Case preservation
		{"upper stays upper",      "HELLO",  3,  "encode", "KHOOR"},
		{"lower stays lower",      "hello",  3,  "encode", "khoor"},
		{"mixed case preserved",   "HeLLo",  3,  "encode", "KhOOr"},

		// Empty input
		{"empty string",           "",       13, "encode", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := postJSON(t, CaesarCipher, buildCipherBody(tc.input, "", tc.shift, tc.mode))
			assertNoError(t, res)
			assertStr(t, res, "output", tc.want)
		})
	}
}

/* ----------------------------------------------------------------
   CaesarCipher — roundtrip: encode then decode must return original
   ---------------------------------------------------------------- */

func TestCaesarCipher_Roundtrip(t *testing.T) {
	cases := []struct {
		name  string
		input string
		key   string
		shift int
	}{
		{"shift-3, no key",          "Hello, World!",           "",      3},
		{"ROT13, no key",            "The quick brown fox",     "",      13},
		{"large shift, no key",      "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "",   7},
		{"with text key KEY",        "Hello, World!",           "KEY",   0},
		{"with text key SECRET",     "The quick brown fox",     "SECRET", 5},
		{"key + shift combined",     "Hello World 123!",        "PASS",  7},
		{"preserves non-alpha",      "foo@bar.com / test=1",    "CIPHER", 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encRes := postJSON(t, CaesarCipher, buildCipherBody(tc.input, tc.key, tc.shift, "encode"))
			assertNoError(t, encRes)
			encoded, _ := encRes["output"].(string)

			decRes := postJSON(t, CaesarCipher, buildCipherBody(encoded, tc.key, tc.shift, "decode"))
			assertNoError(t, decRes)
			decoded, _ := decRes["output"].(string)

			if decoded != tc.input {
				t.Errorf("roundtrip(%q, key=%q, shift=%d):\n  encoded=%q\n  decoded=%q",
					tc.input, tc.key, tc.shift, encoded, decoded)
			}
		})
	}
}

/* ----------------------------------------------------------------
   CaesarCipher — text key specific behaviour
   ---------------------------------------------------------------- */

func TestCaesarCipher_TextKey(t *testing.T) {
	// With key="KEY" (K=10, E=4, Y=24 cycling) and shift=0:
	// H(7)+K(10)=17=R  e(4)+E(4)=8=i  l(11)+Y(24)=35%26=9=j
	// l(11)+K(10)=21=v  o(14)+E(4)=18=s
	t.Run("key=KEY shift=0 encode Hello", func(t *testing.T) {
		res := postJSON(t, CaesarCipher, buildCipherBody("Hello", "KEY", 0, "encode"))
		assertNoError(t, res)
		assertStr(t, res, "output", "Rijvs")
	})

	t.Run("key=KEY shift=0 decode Rijvs", func(t *testing.T) {
		res := postJSON(t, CaesarCipher, buildCipherBody("Rijvs", "KEY", 0, "decode"))
		assertNoError(t, res)
		assertStr(t, res, "output", "Hello")
	})

	// Non-alpha in key should be stripped (key "K-E-Y" == key "KEY")
	t.Run("non-alpha chars in key are stripped", func(t *testing.T) {
		res1 := postJSON(t, CaesarCipher, buildCipherBody("Hello", "KEY",   0, "encode"))
		res2 := postJSON(t, CaesarCipher, buildCipherBody("Hello", "K-E-Y", 0, "encode"))
		out1, _ := res1["output"].(string)
		out2, _ := res2["output"].(string)
		if out1 != out2 {
			t.Errorf("non-alpha key chars should be stripped; KEY=%q vs K-E-Y=%q", out1, out2)
		}
	})

	// Empty key acts as shift-only (neutral 'a' key contributes 0 shift)
	t.Run("empty key = pure shift", func(t *testing.T) {
		res1 := postJSON(t, CaesarCipher, buildCipherBody("Hello", "",  3, "encode"))
		res2 := postJSON(t, CaesarCipher, buildCipherBody("Hello", "a", 3, "encode"))
		out1, _ := res1["output"].(string)
		out2, _ := res2["output"].(string)
		if out1 != out2 {
			t.Errorf("empty key should equal key='a'; got %q vs %q", out1, out2)
		}
	})
}

/* ----------------------------------------------------------------
   CaesarCipher — wrong HTTP method
   ---------------------------------------------------------------- */

func TestCaesarCipher_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	CaesarCipher(w, req)
	var res map[string]any
	decodeInto(t, w, &res)
	assertError(t, res)
}

/* ----------------------------------------------------------------
   helper
   ---------------------------------------------------------------- */

func buildCipherBody(input, key string, shift int, mode string) string {
	return `{"input":` + jsonStr(input) +
		`,"key":` + jsonStr(key) +
		`,"shift":` + itoa(shift) +
		`,"mode":` + jsonStr(mode) + `}`
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
