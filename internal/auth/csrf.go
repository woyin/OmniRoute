package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

const csrfTokenContext = "omniroute-dashboard-csrf-v1"

func CSRFHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		cookie, err := r.Cookie("auth_token")
		secret := GetJWTSecret()
		if err != nil || secret == "" || !validateSessionToken(cookie.Value) {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"token": nil, "expiresAt": nil})
			return
		}

		expires := time.Now().Add(10 * time.Minute).Truncate(time.Second)
		hash := sha256.Sum256([]byte(cookie.Value))
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write([]byte(csrfTokenContext + "\n" + strconv.FormatInt(expires.Unix(), 10) + "\n" + base64.RawURLEncoding.EncodeToString(hash[:])))
		token := "v1." + strconv.FormatInt(expires.Unix(), 10) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		_ = json.NewEncoder(w).Encode(map[string]string{"token": token, "expiresAt": expires.UTC().Format("2006-01-02T15:04:05.000Z")})
	}
}
