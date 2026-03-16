package main

import (
	"log"
	"net/http"
	"os"

	"tools.grejo.in/handler"
)

func main() {
	mux := http.NewServeMux()

	// JSON tools
	mux.HandleFunc("/api/json/format", handler.CORS(handler.JSONFormat))
	mux.HandleFunc("/api/json/validate", handler.CORS(handler.JSONValidate))
	mux.HandleFunc("/api/json/compare", handler.CORS(handler.JSONCompare))
	mux.HandleFunc("/api/json/sort-keys", handler.CORS(handler.JSONSortKeys))

	// URL tools
	mux.HandleFunc("/api/url/parse", handler.CORS(handler.URLParse))
	mux.HandleFunc("/api/url/encode", handler.CORS(handler.URLEncode))
	mux.HandleFunc("/api/url/decode", handler.CORS(handler.URLDecode))

	// Cipher
	mux.HandleFunc("/api/cipher/caesar", handler.CORS(handler.CaesarCipher))

	// Static files — must be last
	mux.Handle("/", http.FileServer(http.Dir("static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8922"
	}

	log.Printf("Listening on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
