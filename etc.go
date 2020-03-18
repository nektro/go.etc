package etc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/aymerick/raymond"
	"github.com/mitchellh/go-homedir"
	"github.com/nektro/go-util/arrays/stringsu"
	"github.com/nektro/go-util/types"
	"github.com/nektro/go-util/util"
	dbstorage "github.com/nektro/go.dbstorage"
	oauth2 "github.com/nektro/go.oauth2"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/pflag"

	. "github.com/nektro/go-util/alias"
)

var (
	AppID      string
	MFS        = new(types.MultiplexFileSystem)
	Database   dbstorage.Database
	ConfigPath string
	Router     *mux.Router
)

var (
	defProviders = []string{}
	appconfFlags = map[string]*string{}
	appFlagTheme []string
)

func PreInit() {
	PreInitAuth()
	PreInitThemes()

	homedir, _ := homedir.Dir()
	pflag.StringVar(&ConfigPath, "config", homedir+"/.config/"+AppID+"/config.json", "")

	pflag.Parse()
}

func Init(appId string, config interface{}, doneURL string, saveOA2Info oauth2.SaveInfoFunc) {
	dRoot := DataRoot()
	util.Log("Reading configuration from:", ConfigPath)

	//
	if !util.DoesDirectoryExist(dRoot) {
		os.MkdirAll(dRoot, os.ModePerm)
	}
	if !util.DoesFileExist(ConfigPath) {
		ioutil.WriteFile(ConfigPath, []byte("{\n}\n"), os.ModePerm)
	}
	InitConfig(ConfigPath, &config)
	pflag.Parse()

	//
	SetSessionName("session_" + appId)

	//
	Database = dbstorage.ConnectSqlite(dRoot + "/access.db")

	//
	Router = mux.NewRouter()

	//
	v := reflect.ValueOf(config).Elem().Elem()
	t := v.Type()

	f, ok := t.FieldByName("Themes")
	if ok {
		themes := []string{}
		themes = append(themes, v.FieldByName(f.Name).Interface().([]string)...)
		themes = append(themes, appFlagTheme...)
		themes = stringsu.Depupe(themes)

		for _, item := range themes {
			loc := dRoot + "/themes/" + item
			util.Log("add-theme:", item)
			util.DieOnError(util.Assert(util.DoesDirectoryExist(loc), F("'%s' directory does not exist!", loc)))
			MFS.Add(http.Dir(loc))
		}
		v.FieldByName(f.Name).Set(reflect.ValueOf(themes))
	}

	//
	// https://github.com/labstack/echo/issues/1038#issuecomment-410294904
	mime.AddExtensionType(".js", "application/javascript")

	//
	MFS.Add(http.Dir("./www/"))

	statikFS, err := fs.New()
	if err == nil {
		MFS.Add(http.FileSystem(statikFS))
	}

	f, ok = t.FieldByName("Providers")
	if ok {
		for _, item := range v.FieldByName(f.Name).Interface().([]oauth2.Provider) {
			util.Log(1, item)
			oauth2.ProviderIDMap[item.ID] = item
		}
	}

	f, ok = t.FieldByName("Clients")
	if ok {
		clients := []oauth2.AppConf{}
		for _, n := range defProviders {
			i := "auth-" + n + "-id"
			s := "auth-" + n + "-secret"
			iv := *appconfFlags[i]
			sv := *appconfFlags[s]
			if len(iv) > 0 && len(sv) > 0 {
				clients = append(clients, oauth2.AppConf{For: n, ID: iv, Secret: sv})
			}
		}
		clients = append(clients, v.FieldByName(f.Name).Interface().([]oauth2.AppConf)...)
		Router.HandleFunc("/login", oauth2.HandleMultiOAuthLogin(helperIsLoggedIn, doneURL, clients))
		Router.HandleFunc("/callback", oauth2.HandleMultiOAuthCallback(doneURL, clients, saveOA2Info))
		v.FieldByName(f.Name).Set(reflect.ValueOf(clients))
	}
}

func DataRoot() string {
	return filepath.Dir(ConfigPath)
}

func helperIsLoggedIn(r *http.Request) bool {
	sess := GetSession(r)
	_, ok := sess.Values["user"]
	return ok
}

func WriteHandlebarsFile(r *http.Request, w http.ResponseWriter, path string, context map[string]interface{}) {
	reader, _ := MFS.Open(path)
	bytes, _ := ioutil.ReadAll(reader)
	template := string(bytes)
	var contentType string
	var result string

	switch r.Header.Get("accept") {
	case "application/json":
		contentType = "application/json"
		resultB, _ := json.Marshal(context)
		result = string(resultB)
	default:
		contentType = "text/html"
		result, _ = raymond.Render(template, context)
	}

	w.Header().Add("Content-Type", contentType)
	fmt.Fprintln(w, result)
}

func AssertPostFormValuesExist(r *http.Request, args ...string) error {
	for _, item := range args {
		v, o := r.PostForm[item]
		if !o {
			return E(F("form[%s] not sent", item))
		}
		if len(v) == 0 {
			return E(F("form[%s] empty", item))
		}
	}
	return nil
}

func WriteResponse(r *http.Request, w http.ResponseWriter, title string, messages ...string) {
	WriteHandlebarsFile(r, w, "/response.hbs", map[string]interface{}{
		"title":    title,
		"messages": messages,
	})
}

func WriteLinkResponse(r *http.Request, w http.ResponseWriter, title string, linkText string, href string, messages ...string) {
	messages = append(messages, "<a href=\""+href+"\">"+linkText+"</a>")
	WriteResponse(r, w, title, messages...)
}

func StartServer(port int) {
	util.DieOnError(util.Assert(util.IsPortAvailable(port), F("Binding to port %d failed.", port)), "It may be taken or you may not have permission to. Aborting!")
	p := strconv.Itoa(port)
	util.Log("Initialization complete. Starting server on port " + p)
	Router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(MFS)))
	Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-Frame-Options", "sameorigin")
			w.Header().Add("X-Content-Type-Options", "nosniff")
			w.Header().Add("Referrer-Policy", "origin")
			next.ServeHTTP(w, r)
		})
	})
	srv := &http.Server{
		Handler:      Router,
		Addr:         ":" + p,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.ListenAndServe()
}

// FixBareVersion will convert a 'vMASTER' version string to a string
// similar to 'vMASTER-2020.02.12-6cae79d'. Always append go version.
func FixBareVersion(vs string) string {
	if vs == "vMASTER" {
		// add date
		pathS, _ := filepath.Abs(os.Args[0])
		s, _ := os.Stat(pathS)
		vs += "-" + strings.ReplaceAll(s.ModTime().UTC().String()[:10], "-", ".")

		// add git hash
		b, _ := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
		if len(b) == 8 {
			vs += "-" + string(b)[:7]
		}
	}
	vs += "-" + runtime.Version()
	return vs
}
