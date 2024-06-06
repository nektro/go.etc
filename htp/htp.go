package htp

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/nektro/go-util/util"
	"github.com/nektro/go-util/vflag"
)

// globals
var (
	router          *mux.Router
	ErrorHandleFunc func(http.ResponseWriter, *http.Request, string)
	controllers     map[*http.Request]*Controller
	base            = vflag.String("base", "/", "The path to mount all listeners on")
	baseReal        string
	srv             *http.Server
	mtx             = new(sync.Mutex)
	allowedips      = []string{}
)

func init() {
	// fix mime type handling
	associations := [][2]string{
		{".aac", "audio/aac"},
		{".abw", "application/x-abiword"},
		{".arc", "application/x-freearc"},
		{".avi", "video/x-msvideo"},
		{".azw", "application/vnd.amazon.ebook"},
		{".bin", "application/octet-stream"},
		{".bmp", "image/bmp"},
		{".bz", "application/x-bzip"},
		{".bz2", "application/x-bzip2"},
		{".cda", "application/x-cdf"},
		{".csh", "application/x-csh"},
		{".css", "text/css"},
		{".csv", "text/csv"},
		{".doc", "application/msword"},
		{".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{".eot", "application/vnd.ms-fontobject"},
		{".epub", "application/epub+zip"},
		{".gz", "application/gzip"},
		{".gif", "image/gif"},
		{".htm", "text/html"},
		{".html", "text/html"},
		{".ico", "image/vnd.microsoft.icon"},
		{".ics", "text/calendar"},
		{".jar", "application/java-archive"},
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".js", "application/javascript"},
		{".json", "application/json"},
		{".jsonld", "application/ld+json"},
		{".mid", "audio/midi"},
		{".mjs", "application/javascript"},
		{".mp3", "audio/mpeg"},
		{".mp4", "video/mp4"},
		{".mpeg", "video/mpeg"},
		{".mpkg", "application/vnd.apple.installer+xml"},
		{".odp", "application/vnd.oasis.opendocument.presentation"},
		{".ods", "application/vnd.oasis.opendocument.spreadsheet"},
		{".odt", "application/vnd.oasis.opendocument.text"},
		{".oga", "audio/ogg"},
		{".ogv", "video/ogg"},
		{".ogx", "application/ogg"},
		{".opus", "audio/opus"},
		{".otf", "font/otf"},
		{".png", "image/png"},
		{".pdf", "application/pdf"},
		{".php", "application/x-httpd-php"},
		{".ppt", "application/vnd.ms-powerpoint"},
		{".pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{".rar", "application/vnd.rar"},
		{".rtf", "application/rtf"},
		{".sh", "application/x-sh"},
		{".svg", "image/svg+xml"},
		{".swf", "application/x-shockwave-flash"},
		{".tar", "application/x-tar"},
		{".tiff", "image/tiff"},
		{".ts", "video/mp2t"},
		{".ttf", "font/ttf"},
		{".txt", "text/plain"},
		{".vsd", "application/vnd.visio"},
		{".wav", "audio/wav"},
		{".weba", "audio/webm"},
		{".webm", "video/webm"},
		{".webp", "image/webp"},
		{".woff", "font/woff"},
		{".woff2", "font/woff2"},
		{".xhtml", "application/xhtml+xml"},
		{".xls", "application/vnd.ms-excel"},
		{".xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{".xml", "application/xml"},
		{".xul", "application/vnd.mozilla.xul+xml"},
		{".zip", "application/zip"},
		{".3gp", "video/3gpp"},
		{".3g2", "video/3gpp2"},
		{".7z", "application/x-7z-compressed"},
	}
	for _, item := range associations {
		mime.AddExtensionType(item[0], item[1])
	}
}

// PreInit sets up flags
func PreInit() {
	vflag.StringArrayVar(&allowedips, "allow-ip", []string{}, "Only allow requests from specific IP pattern. Use 'x' for replacements.")
}

// Init sets up globals to their default state
func Init() {
	router = mux.NewRouter()
	ErrorHandleFunc = func(http.ResponseWriter, *http.Request, string) {}
	controllers = map[*http.Request]*Controller{}
	util.DieOnError(util.Assert(strings.HasSuffix(*base, "/"), "--base must end in '/'"))
	baseReal = strings.TrimSuffix(*base, "/")

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Frame-Options", "sameorigin")
			w.Header().Add("X-Content-Type-Options", "nosniff")
			w.Header().Add("Referrer-Policy", "origin")
			next.ServeHTTP(w, r)
		})
	})
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a := strings.Split(strings.Split(r.RemoteAddr, ":")[0], ".")
			for _, item := range allowedips {
				for j, jtem := range strings.Split(item, ".") {
					if jtem == "x" {
						continue
					}
					if jtem != a[j] {
						fmt.Fprintln(w, "403 forbidden")
						fmt.Fprintln(w, "ip not allowed")
						return
					}
				}
			}
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
		rt.PathPrefix(baseReal + strings.TrimSuffix(path, "*"))
	} else {
		rt.Path(baseReal + path)
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
				panic(rcvs)
			}
		}()
		h(w, r)
		mtx.Lock()
		delete(controllers, r)
		mtx.Unlock()
	})
}

// GetController allows you to gain access to this method's htp.Controller
func GetController(r *http.Request) *Controller {
	mtx.Lock()
	defer mtx.Unlock()

	c, ok := controllers[r]
	if !ok {
		c = &Controller{r}
		controllers[r] = c
	}
	return c
}

// RegisterFileSystem is a custom version of Register where it adds a http.FileSystem to the router
func RegisterFileSystem(fs http.FileSystem) {
	p := baseReal + "/"
	router.PathPrefix(p).Handler(http.StripPrefix(p, http.FileServer(fs)))
}

// Base returns the root that all methods are mounted on
func Base() string {
	return baseReal + "/"
}

// StartServer initializes this server and listens on port
func StartServer(bind string, port int) {
	p := strconv.Itoa(port)
	util.DieOnError(util.Assert(util.IsPortAvailable(port), "Binding to port "+p+" failed."), "It may be taken or you may not have permission to. Aborting!")
	util.Log("Starting server on port " + p)
	if util.AreWeInContainer() {
		util.LogWarn("Looks like we might be running inside a container, so " + p + " might not be the actual port to access this server.")
	}
	util.Log("Initialization complete.")
	srv = &http.Server{
		Handler: router,
		Addr:    bind + ":" + p,
	}
	srv.ListenAndServe()
}

// StopServer performs a graceful shutdown of the HTTP server
func StopServer() {
	srv.Close()
	srv.Shutdown(context.Background())
}
