package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
)

// URLParse breaks a URL into its constituent parts.
func URLParse(w http.ResponseWriter, r *http.Request) {
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
	u, err := url.Parse(req.Input)
	if err != nil {
		writeError(w, fmt.Sprintf("invalid URL: %s", err.Error()))
		return
	}

	type param struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	// Parse query string into sorted key-value pairs.
	queryMap, _ := url.ParseQuery(u.RawQuery)
	keys := make([]string, 0, len(queryMap))
	for k := range queryMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	queryList := make([]param, 0)
	for _, k := range keys {
		for _, v := range queryMap[k] {
			queryList = append(queryList, param{Key: k, Value: v})
		}
	}

	// Parse fragment into its own query params if present.
	fragmentQuery := make([]param, 0)
	if u.Fragment != "" {
		fragMap, ferr := url.ParseQuery(u.Fragment)
		if ferr == nil && len(fragMap) > 1 {
			fkeys := make([]string, 0, len(fragMap))
			for k := range fragMap {
				fkeys = append(fkeys, k)
			}
			sort.Strings(fkeys)
			for _, k := range fkeys {
				for _, v := range fragMap[k] {
					fragmentQuery = append(fragmentQuery, param{Key: k, Value: v})
				}
			}
		}
	}

	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	writeJSON(w, map[string]any{
		"scheme":         u.Scheme,
		"host":           u.Hostname(),
		"port":           u.Port(),
		"path":           u.Path,
		"query":          queryList,
		"fragment":       u.Fragment,
		"fragment_query": fragmentQuery,
		"raw_query":      u.RawQuery,
		"username":       username,
		"password":       password,
	})
}

// URLEncode percent-encodes a string suitable for use as a query parameter value.
func URLEncode(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, map[string]string{"output": url.QueryEscape(req.Input)})
}

// URLDecode decodes a percent-encoded string.
func URLDecode(w http.ResponseWriter, r *http.Request) {
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
	decoded, err := url.QueryUnescape(req.Input)
	if err != nil {
		writeError(w, fmt.Sprintf("decode error: %s", err.Error()))
		return
	}
	writeJSON(w, map[string]string{"output": decoded})
}
