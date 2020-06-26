package etc

import (
	"net/http"
	"os"
	"time"

	"github.com/nektro/go.etc/htp"
	"github.com/nektro/go.etc/jwt"
)

// JWTSet sets a 'jwt' cookie on the provided ResponseWriter
func JWTSet(w http.ResponseWriter, sub string) {
	n, _ := os.Hostname()
	http.SetCookie(w, &http.Cookie{
		Name:   "jwt",
		Value:  jwt.Get("astheno."+AppID+"."+Version+"."+n, sub, JWTSecret, Epoch, time.Hour*24*30),
		MaxAge: 0,
	})
}

// JWTGetClaims reads the 'jwt' cookie and returns claims within it, if they are valid
func JWTGetClaims(c *htp.Controller, r *http.Request) jwt.MapClaims {
	clms, err := jwt.VerifyRequest(r, JWTSecret)
	c.Assert(err == nil, "403: "+err.Error())
	return clms
}

// JWTDestroy tells the ResponseWriter to delete the 'jwt' cookie
func JWTDestroy(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "jwt",
		MaxAge: -1,
	})
}
