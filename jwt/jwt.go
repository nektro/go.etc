package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
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
func Verify(token string, secret string) (MapClaims, error) {
	tokenO, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method: " + fmt.Sprintf("%v", t.Header["alg"]))
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, errors.New("astheno/jwt: token: " + err.Error())
	}
	if !tokenO.Valid {
		return nil, errors.New("astheno/jwt: token: not valid")
	}
	claims, ok := tokenO.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("astheno/jwt: claims: not a MapClaims")
	}
	if err = claims.Valid(); err != nil {
		return nil, errors.New("astheno/jwt: claims: " + err.Error())
	}
	return MapClaims(claims), nil
}

// VerifyRequest is a shortcut for `Verify(FromRequest(r), secret)`
func VerifyRequest(r *http.Request, secret string) (MapClaims, error) {
	return Verify(FromRequest(r), secret)
}
