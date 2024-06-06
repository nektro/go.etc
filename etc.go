package etc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/mitchellh/go-homedir"
	"github.com/nektro/go-util/arrays/stringsu"
	"github.com/nektro/go-util/types"
	"github.com/nektro/go-util/util"
	"github.com/nektro/go-util/vflag"
	dbstorage "github.com/nektro/go.dbstorage"
	"github.com/nektro/go.etc/htp"
	"github.com/nektro/go.etc/internal"
	"github.com/nektro/go.etc/jwt"
	oauth2 "github.com/nektro/go.oauth2"
	"github.com/rakyll/statik/fs"

	. "github.com/nektro/go-util/alias"
)

// globals
var (
	AppID      string
	Version    = "vMASTER"
	MFS        = new(types.MultiplexFileSystem)
	Database   dbstorage.Database
	ConfigPath string
	JWTSecret  string
	Bind       string
	Port       int
	Epoch      = internal.Epoch
	HtpErrCb   = func(r *http.Request, w http.ResponseWriter, good bool, status int, message string) {}
)

var (
	defProviders = []string{}
	appFlagTheme []string
	homedirV, _  = homedir.Dir()
)

// PreInit registers and parses application flags
func PreInit() {
	vflag.StringArrayVar(&appFlagTheme, "theme", []string{}, "A CLI way to add config themes.")
	vflag.StringVar(&ConfigPath, "config", homedirV+"/.config/"+AppID+"/config.json", "")
	vflag.StringVar(&JWTSecret, "jwt-secret", util.RandomString(64), "Private secret to sign and verify JWT auth tokens with.")
	vflag.StringVar(&Bind, "bind", "127.0.0.1", "IP to bind")
	vflag.IntVar(&Port, "port", 8000, "The port to bind the web server to.")
	htp.PreInit()

	vflag.Parse()
	SetSessionName("session_" + AppID)
}

// Init sets up app-agnostic features
func Init(config interface{}, doneURL string, saveOA2Info oauth2.SaveInfoFunc) {
	dRoot := DataRoot()

	//
	if !util.DoesDirectoryExist(dRoot) {
		os.MkdirAll(dRoot, os.ModePerm)
	}
	if !util.DoesFileExist(ConfigPath) {
		ioutil.WriteFile(ConfigPath, []byte("{\n}\n"), os.ModePerm)
	}
	InitConfig(ConfigPath, &config)
	vflag.Parse()

	//
	db, err := connectDB()
	util.DieOnError(err)
	util.Log("etc:", "db:", db.DriverName())
	Database = db

	//
	htp.Init()

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
		clients = append(clients, v.FieldByName(f.Name).Interface().([]oauth2.AppConf)...)
		callbackPath := htp.Base() + "callback"
		loginH, callbackH := oauth2.GetHandlers(helperIsLoggedIn, doneURL, callbackPath, &clients, saveOA2Info)
		v.FieldByName(f.Name).Set(reflect.ValueOf(clients))
		htp.Register("/login", "GET", loginH)
		htp.Register("/callback", "GET", callbackH)
		v.FieldByName(f.Name).Set(reflect.ValueOf(clients))
	}

	htp.ErrorHandleFunc = func(w http.ResponseWriter, r *http.Request, data string) {
		code, _ := strconv.Atoi(data[:3])
		good := !(code >= 400)
		HtpErrCb(r, w, good, int(code), data[5:])
	}
}

func connectDB() (dbstorage.Database, error) {
	if len(os.Getenv("POSTGRES_URL")) > 0 {
		return dbstorage.ConnectPostgres()
	}
	if len(os.Getenv("MYSQL_URL")) > 0 {
		return dbstorage.ConnectMysql()
	}
	return dbstorage.ConnectSqlite(DataRoot() + "/access.db")
}

func DataRoot() string {
	return filepath.Dir(ConfigPath)
}

func helperIsLoggedIn(r *http.Request) bool {
	_, err := jwt.VerifyRequest(r, JWTSecret)
	return err == nil
}

func WriteHandlebarsFile(r *http.Request, w http.ResponseWriter, path string, context map[string]interface{}) {
	{
		al, ok := r.Header["Accept-Language"]
		if !ok || al == nil {
			al = []string{}
		}
		arr := []string{}
		for _, item := range al {
			arr = append(arr, strings.Split(item, ";")[0])
		}
		context["languages"] = strings.Join(arr, ",")
	}
	reader, _ := MFS.Open(path)
	bytes, _ := ioutil.ReadAll(reader)
	template := string(bytes)
	result, _ := raymond.Render(template, context)
	fmt.Fprintln(w, result)
}

func StartServer() {
	htp.RegisterFileSystem(MFS)
	htp.StartServer(Bind, Port)
}

// FixBareVersion will convert a 'vMASTER' version string to a string
// similar to 'vMASTER-2020.02.12-6cae79d'. Always append go version.
func FixBareVersion() {
	if Version == "vMASTER" {
		// add date
		pathS, _ := filepath.Abs(os.Args[0])
		s, _ := os.Stat(pathS)
		Version += "-" + strings.ReplaceAll(s.ModTime().UTC().String()[:10], "-", ".")

		// add git hash
		b, _ := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
		if len(b) == 8 {
			Version += "-" + string(b)[:7]
		}
	}
	Version += "-" + runtime.Version()
}
