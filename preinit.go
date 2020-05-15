package etc

import (
	"strings"

	"github.com/nektro/go-util/vflag"
	oauth2 "github.com/nektro/go.oauth2"
)

func PreInitAuth() {
	for k, _ := range oauth2.ProviderIDMap {
		n := strings.ReplaceAll(strings.ReplaceAll(k, "_", "-"), ".", "-")
		defProviders = append(defProviders, n)
		i := "auth-" + n + "-id"
		appconfFlags[i] = vflag.String(i, "", "Client ID for "+k+" OAuth2 authentication.")
		s := "auth-" + n + "-secret"
		appconfFlags[s] = vflag.String(s, "", "Client Secret for "+k+" OAuth2 authentication.")
	}
}

func PreInitThemes() {
	vflag.StringArrayVar(&appFlagTheme, "theme", []string{}, "A CLI way to add config themes.")
}
