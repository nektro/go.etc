package etc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/mitchellh/go-homedir"
	"github.com/nektro/go-util/types"
	"github.com/nektro/go-util/util"
	dbstorage "github.com/nektro/go.dbstorage"
	oauth2 "github.com/nektro/go.oauth2"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/pflag"

	. "github.com/nektro/go-util/alias"
)

var (
	MFS      = new(types.MultiplexFileSystem)
	Database dbstorage.Database
)

var (
	appconfFlags = map[string]*string{}
	appFlagTheme []string
)

func PreInit() {
	//
	for k, _ := range oauth2.ProviderIDMap {
		n := strings.ReplaceAll(strings.ReplaceAll(k, "_", "-"), ".", "-")
		i := "auth-" + n + "-id"
		appconfFlags[i] = pflag.String(i, "", "Client ID for "+k+" OAuth2 authentication.")
		s := "auth-" + n + "-secret"
		appconfFlags[s] = pflag.String(s, "", "Client Secret for "+k+" OAuth2 authentication.")
	}

	//
	pflag.StringArrayVar(&appFlagTheme, "theme", []string{}, "A CLI way to add config themes.")

	//
	pflag.Parse()
}

func Init(appId string, config interface{}, doneURL string, saveOA2Info oauth2.SaveInfoFunc) {
	homedir, _ := homedir.Dir()
	dataRoot := path.Join(homedir, ".config", appId)
	configPath := path.Join(dataRoot, "config.json")
	util.Log("Reading configuration from", configPath)

	//
	if !util.DoesDirectoryExist(dataRoot) {
		os.MkdirAll(dataRoot, os.ModePerm)
	}
	if !util.DoesFileExist(configPath) {
		ioutil.WriteFile(configPath, []byte("{\n}\n"), os.ModePerm)
	}
	InitConfig(configPath, &config)

	//
	SetSessionName("session_" + appId)

	//
	Database = dbstorage.ConnectSqlite(dataRoot + "/access.db")

	//
	v := reflect.ValueOf(config).Elem().Elem()
	t := v.Type()

	f, ok := t.FieldByName("Themes")
	if ok {
		themes := []string{}
		themes = append(themes, v.FieldByName(f.Name).Interface().([]string)...)
		themes = append(themes, appFlagTheme...)

		for _, item := range themes {
			loc := dataRoot + "/themes/" + item
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

	//
	http.HandleFunc("/", http.FileServer(MFS).ServeHTTP)

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
		for k := range oauth2.ProviderIDMap {
			n := strings.ReplaceAll(strings.ReplaceAll(k, "_", "-"), ".", "-")
			i := "auth-" + n + "-id"
			s := "auth-" + n + "-secret"
			iv := *appconfFlags[i]
			sv := *appconfFlags[s]
			if len(iv) > 0 && len(sv) > 0 {
				clients = append(clients, oauth2.AppConf{For: n, ID: iv, Secret: sv})
			}
		}
		clients = append(clients, v.FieldByName(f.Name).Interface().([]oauth2.AppConf)...)
		util.DieOnError(util.Assert(len(clients) > 0, "Must provide at least 1 oauth2.AppConf in the 'Clients' config array."))
		http.HandleFunc("/login", oauth2.HandleMultiOAuthLogin(helperIsLoggedIn, doneURL, clients))
		http.HandleFunc("/callback", oauth2.HandleMultiOAuthCallback(doneURL, clients, saveOA2Info))
		v.FieldByName(f.Name).Set(reflect.ValueOf(clients))
	}
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
	http.ListenAndServe(":"+p, nil)
}
