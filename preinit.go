package etc

import (
	"github.com/nektro/go-util/vflag"
)

func PreInitThemes() {
	vflag.StringArrayVar(&appFlagTheme, "theme", []string{}, "A CLI way to add config themes.")
}
