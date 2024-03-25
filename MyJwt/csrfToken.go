package MyJwt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
)

// GenerateCSRFToken generates a random token for CSRF protection
func GenerateRandomString() (string, error) {
	b := make([]byte, 32) // 32 bytes generates a token of sufficient randomness
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func SetCSRFCookie(w http.ResponseWriter, csrfString string) {
	cookie := &http.Cookie{
		Name:     CSRF_TOKEN_NAME,
		Value:    csrfString,
		HttpOnly: true,
		Secure:   true,
		Path:     "/", // Ensure the path matches the one used when setting the cookie
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

func GetCSRFTokenFromReq(r *http.Request) (csrf string) {
	csrf = r.FormValue(CSRF_TOKEN_NAME)
	if csrf != "" {
		return
	}
	return r.Header.Get(CSRF_TOKEN_NAME)
}

func CompareCSRFToken(r *http.Request, csrfToken string) bool {
	if csrfToken == "" {
		return false
	}
	return csrfToken == GetCSRFTokenFromReq(r)
}

func ValidateCSRFToken(r *http.Request) bool {
	csrfCookie, err := r.Cookie(CSRF_TOKEN_NAME)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return CompareCSRFToken(r, csrfCookie.Value)
}

func RemoveCSRFCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     CSRF_TOKEN_NAME,
		Value:    "",
		HttpOnly: true,
		Path:     "/", // Ensure the path matches the one used when setting the cookie
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}
