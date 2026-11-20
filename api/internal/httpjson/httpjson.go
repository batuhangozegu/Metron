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
	value := map[string]string{"error": message}
	switch status {
	case 400:
		value["code"] = "invalid_input"
	case 401:
		value["code"] = "authentication_required"
	case 415:
		value["code"] = "unsupported_media_type"
	default:
		if status >= 500 {
			value["code"] = "unavailable"
		}
	}
	// Shared request/session failures also appear on the bilingual risk screen.
	translations := map[string]string{
		"market must be spot or futures":                                              "Piyasa spot veya futures olmalıdır",
		"name must contain 1 to 100 characters":                                       "Ad 1 ile 100 karakter arasında olmalıdır",
		"paper_balance must be a decimal string with at most 8 decimal places":        "Kağıt hesap bakiyesi en fazla 8 ondalık basamaklı bir dize olmalıdır",
		"paper_balance must be greater than 0 and at most 10000000":                   "Kağıt hesap bakiyesi 0'dan büyük ve en fazla 10000000 olmalıdır",
		"accounts temporarily unavailable":                                            "Hesaplar geçici olarak kullanılamıyor",
		"invalid JSON body":                                                           "Geçersiz JSON gövdesi",
		"expected one JSON object":                                                    "Tek bir JSON nesnesi bekleniyor",
		"Content-Type must be application/json":                                       "Content-Type application/json olmalıdır",
		"authentication required":                                                     "Oturum açmanız gerekiyor",
		"authentication temporarily unavailable":                                      "Kimlik doğrulama geçici olarak kullanılamıyor",
		"authentication unavailable: configure APP_ENCRYPTION_KEY (32 bytes, base64)": "Kimlik doğrulama kullanılamıyor: APP_ENCRYPTION_KEY ayarlayın (32 bayt, base64)",
		"authentication unavailable: configure DATABASE_URL and run migrations":       "Kimlik doğrulama kullanılamıyor: DATABASE_URL ayarlayın ve veritabanı geçişlerini çalıştırın",
	}
	if tr, ok := translations[message]; ok {
		value["error_tr"] = tr
	}
	if writer, ok := w.(interface{ WriteError(int, string, string) }); ok {
		writer.WriteError(status, message, value["error_tr"])
		return
	}
	Write(w, status, value)
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
