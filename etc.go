package etc

import (
	"fmt"
	"net/http"

	"github.com/aymerick/raymond"
	"github.com/nektro/go-util/util"

)

func WriteHandlebarsFile(r *http.Request, w http.ResponseWriter, path string, context map[string]interface{}) {
	template := string(util.ReadFile(path))
	result, _ := raymond.Render(template, context)
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintln(w, result)
}
