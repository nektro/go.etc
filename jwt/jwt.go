package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Get takes in init values and returns a signed HS256 JWT as a string.
func Get(iss string, sub string, secret string, nbf time.Time, exp time.Duration) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": iss,
		"sub": sub,
		"exp": time.Now().Add(exp).Unix(),
		"nbf": nbf.Unix(),
		"iat": time.Now().Unix(),
	})
	tokenS, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}
	return tokenS
}

// MapClaims represents the json data of this JWT
type MapClaims jwt.MapClaims

// Verify takes in a JWT string an a corresponding HS256 secret and verifies the token and its claims are valid.
func Verify(token string, secret string) (MapClaims, bool) {
	tokenO, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method: " + fmt.Sprintf("%v", t.Header["alg"]))
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, false
	}
	if !tokenO.Valid {
		return nil, false
	}
	claims, ok := tokenO.Claims.(jwt.MapClaims)
	if !ok {
		return nil, false
	}
	if claims.Valid() != nil {
		return nil, false
	}
	return MapClaims(claims), true
}

// VerifyRequest is a shortcut for `Verify(FromRequest(r), secret)`
func VerifyRequest(r *http.Request, secret string) (MapClaims, bool) {
	return Verify(FromRequest(r), secret)
}
