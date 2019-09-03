package etc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aymerick/raymond"
	"github.com/mitchellh/go-homedir"
	"github.com/nektro/go-util/types"
	"github.com/nektro/go-util/util"
	dbstorage "github.com/nektro/go.dbstorage"

	. "github.com/nektro/go-util/alias"
)

var (
	MFS      = new(types.MultiplexFileSystem)
	Database *dbstorage.DbProxy
)

func Init(appId string, config interface{}) {
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
	SetSessionName("session_" + appId + "_test")

	//
	Database = dbstorage.ConnectSqlite(dataRoot + "/access.db")
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
