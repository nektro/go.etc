package etc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/aymerick/raymond"
	"github.com/mitchellh/go-homedir"
	"github.com/nektro/go-util/types"
	"github.com/nektro/go-util/util"
	dbstorage "github.com/nektro/go.dbstorage"
	oauth2 "github.com/nektro/go.oauth2"
	"github.com/rakyll/statik/fs"

	. "github.com/nektro/go-util/alias"
)

var (
	MFS      = new(types.MultiplexFileSystem)
	Database dbstorage.Database
)

func Init(appId string, config interface{}, doneURL string, saveOA2Info oauth2.SaveInfoFunc) {
	homedir, _ := homedir.Dir()
	dataRoot := homedir + "/.config/" + appId
	configPath := dataRoot + "/config.json"
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
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == "Themes" {
			for _, item := range v.FieldByName(f.Name).Interface().([]string) {
				loc := dataRoot + "/themes/" + item
				util.Log("[add-theme]", item)
				util.DieOnError(util.Assert(util.DoesDirectoryExist(loc), F("'%s' directory does not exist!", loc)))
				MFS.Add(http.Dir(loc))
			}
		}
		if f.Name == "Providers" {
			for _, item := range v.FieldByName(f.Name).Interface().([]oauth2.Provider) {
				util.Log(1, item)
				oauth2.ProviderIDMap[item.ID] = item
			}
		}
		if f.Name == "Clients" {
			clients := v.FieldByName(f.Name).Interface().([]oauth2.AppConf)
			http.HandleFunc("/login", oauth2.HandleMultiOAuthLogin(helperIsLoggedIn, doneURL, clients))
			http.HandleFunc("/callback", oauth2.HandleMultiOAuthCallback(doneURL, clients, saveOA2Info))
		}
	}

	//
	MFS.Add(http.Dir("./www/"))

	statikFS, err := fs.New()
	util.DieOnError(err)
	MFS.Add(http.FileSystem(statikFS))

	//
	http.HandleFunc("/", http.FileServer(MFS).ServeHTTP)
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
	result, _ := raymond.Render(template, context)
	w.Header().Add("Content-Type", "text/html")
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

func RunOnClose(f func()) {
	gracefulStop := make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		sig := <-gracefulStop
		util.Log(F("Caught signal '%+v'", sig))
		f()
		os.Exit(0)
	}()
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
