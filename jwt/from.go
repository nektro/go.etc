package jwt

import (
	"net/http"
)

// FromRequest reads the given http request to see if it contains a JWT.
func FromRequest(r *http.Request) string {
	bearer := ""
	for _, item := range []func(*http.Request) string{tokenFromHeader, tokenFromCookie, tokenFromQuery} {
		bearer = item(r)
		if len(bearer) > 0 {
			break
		}
	}
	return bearer
}

func tokenFromHeader(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) <= 7 || bearer[0:6] != "Bearer" {
		return ""
	}
	return bearer[7:]
}

func tokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func tokenFromQuery(r *http.Request) string {
	return r.URL.Query().Get("jwt")
}
