package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	etc "github.com/nektro/go.etc"
	"github.com/nektro/go.etc/htp"
	oauth2 "github.com/nektro/go.oauth2"
)

type configT struct {
	Clients   []oauth2.AppConf  `json:"clients"`
	Providers []oauth2.Provider `json:"providers"`
	Themes    []string          `json:"themes"`
}

var (
	configV = new(configT)
)

// Run this with `go run ./example/main.go` with whichever options you with to play around with.
func main() {
	etc.PreInit()

	etc.Init(&configV, "./dashboard", func(w http.ResponseWriter, r *http.Request, provider, id, name string, data map[string]interface{}) {
		log.Println("user-login:", provider, id, name)
		etc.JWTSet(w, provider+"\n"+id)
	})

	etc.HtpErrCb = func(r *http.Request, w http.ResponseWriter, good bool, code int, message string) {
		resp := map[string]interface{}{
			"success": good,
			"message": message,
		}
		w.Header().Add("content-type", "application/json")
		dat, _ := json.Marshal(resp)
		fmt.Fprintln(w, string(dat))
	}

	htp.Register("/dashboard", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		c := htp.GetController(r)
		l := etc.JWTGetClaims(c, r)
		s := l["sub"].(string)
		d := strings.SplitN(s, "\n", 2)
		fmt.Fprintln(w, "provider: ", d[0])
		fmt.Fprintln(w, "snowflake:", d[1])
	})

	etc.StartServer()
}
