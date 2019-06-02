package etc

import (
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var (
	cookie    string
	randomKey = securecookie.GenerateRandomKey(32)
	store     = sessions.NewCookieStore(randomKey)
)

func SetSessionName(name string) {
	cookie = name
}

func GetSession(r *http.Request) *sessions.Session {
	sess, _ := store.Get(r, cookie)
	return sess
}
