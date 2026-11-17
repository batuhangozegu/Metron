// Package httpjson shares JSON request and error handling between API modules.
package httpjson

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

func Write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func Error(w http.ResponseWriter, status int, message string) {
	Write(w, status, map[string]string{"error": message})
}

func Decode(w http.ResponseWriter, r *http.Request, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		Error(w, 400, "invalid JSON body")
		return false
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		Error(w, 400, "expected one JSON object")
		return false
	}
	return true
}

// JSONOnly prevents cross-site HTML forms from making cookie-authenticated
// mutations. CORS must not allow untrusted origins with credentials.
func JSONOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || media != "application/json" {
				Error(w, 415, "Content-Type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
