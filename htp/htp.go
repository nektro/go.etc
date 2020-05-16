package htp

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nektro/go-util/util"

	. "github.com/nektro/go-util/alias"
)

// globals
var (
	router          *mux.Router
	ErrorHandleFunc func(http.ResponseWriter, *http.Request, string)
	controller      *Controller
)

// Init sets up globals to their default state
func Init() {
	router = mux.NewRouter()
	ErrorHandleFunc = func(http.ResponseWriter, *http.Request, string) {}
	controller = &Controller{}

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Frame-Options", "sameorigin")
			w.Header().Add("X-Content-Type-Options", "nosniff")
			w.Header().Add("Referrer-Policy", "origin")
			next.ServeHTTP(w, r)
		})
	})
}

// Register adds a handler to this router.
func Register(path, method string, h func(w http.ResponseWriter, r *http.Request)) {
	methods := []string{}
	if len(method) > 0 {
		methods = append(methods, method)
	}
	if method == http.MethodGet {
		methods = append(methods, http.MethodHead)
	}
	rt := router.NewRoute()
	rt.Methods(methods...)
	if strings.HasSuffix(path, "/*") {
		rt.PathPrefix(strings.TrimSuffix(path, "*"))
	} else {
		rt.Path(path)
	}
	rt.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rcv := recover(); rcv != nil {
				rcvs := fmt.Sprintf("%v", rcv)
				if t1 := strings.TrimPrefix(rcvs, "htp: "); t1 != rcvs {
					if t2 := strings.TrimPrefix(t1, "assertion failed: "); t2 != t1 {
						ErrorHandleFunc(w, r, t2)
						return
					}
					if t2 := strings.TrimPrefix(t1, "redirect: "); t2 != t1 {
						w.Header().Add("Location", t2)
						w.WriteHeader(http.StatusFound)
						return
					}
				}
			}
		}()
		h(w, r)
	})
}

// GetController allows you to gain access to this method's htp.Controller
func GetController(r *http.Request) *Controller {
	return controller
}

// RegisterFileSystem is a custom version of Register where it adds a http.FileSystem to the router
func RegisterFileSystem(fs http.FileSystem) {
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(fs)))
}

// StartServer initializes this server and listens on port
func StartServer(port int) {
	util.DieOnError(util.Assert(util.IsPortAvailable(port), F("Binding to port %d failed.", port)), "It may be taken or you may not have permission to. Aborting!")
	p := strconv.Itoa(port)
	util.Log("Initialization complete. Starting server on port " + p)
	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + p,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.ListenAndServe()
}
