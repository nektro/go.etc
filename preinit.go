package etc

import (
	"strings"

	oauth2 "github.com/nektro/go.oauth2"
	"github.com/spf13/pflag"
)

func PreInitAuth() {
	for k, _ := range oauth2.ProviderIDMap {
		n := strings.ReplaceAll(strings.ReplaceAll(k, "_", "-"), ".", "-")
		i := "auth-" + n + "-id"
		appconfFlags[i] = pflag.String(i, "", "Client ID for "+k+" OAuth2 authentication.")
		s := "auth-" + n + "-secret"
		appconfFlags[s] = pflag.String(s, "", "Client Secret for "+k+" OAuth2 authentication.")
	}
}

func PreInitThemes() {
	pflag.StringArrayVar(&appFlagTheme, "theme", []string{}, "A CLI way to add config themes.")
}
