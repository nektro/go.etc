package etc

import (
	"github.com/nektro/go.hosts"

	. "github.com/nektro/go-util/util"
)

func ReadAllowedHostnames(path string) {
	DieOnError(Assert(DoesFileExist(path), "allowed_domains.txt does not exist!"))
	ReadFileLines(path, func(line string) {
		hosts.Allow(line)
	})
}
