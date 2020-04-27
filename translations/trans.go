package translations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aymerick/raymond"
	"github.com/nektro/go-util/arrays/stringsu"
	"github.com/nektro/go-util/util"
	etc "github.com/nektro/go.etc"
	"github.com/spf13/pflag"
)

// For fetching and integrating Crowdin translations built with:
// https://github.com/nektro/astheno.rocks/blob/master/get_translations.sh

var (
	Server    = "https://astheno.rocks"
	Languages []string
	Words     = map[string]map[string]string{}
	flagSkip  = pflag.Bool("skip-translation-fetch", false, "Enable this flag to only read native translation data.")
)

func get(end string) []byte {
	req, _ := http.NewRequest(http.MethodGet, Server+"/"+etc.AppID+"/translations"+end, nil)
	bys, err := util.DoHttpFetch(req)
	util.DieOnError(err)
	return bys
}

func Fetch() {
	// read translations from astheno.rocks
	util.Log("translations:", "fetching...")
	if !*flagSkip {
		json.Unmarshal(get("/_languages.json"), &Languages)
		fmt.Print("|")
		for _, item := range Languages {
			if len(item) == 0 {
				continue
			}
			mp := map[string]string{}
			json.Unmarshal(get("/"+item+".json"), &mp)
			Words[item] = mp
			fmt.Print("|")
		}
	}
	{
		// add default english values
		fl, _ := etc.MFS.Open("/sources.xml")
		doc, _ := goquery.NewDocumentFromReader(fl)
		mp := map[string]string{}
		doc.Find("resources string").Each(func(_ int, el *goquery.Selection) {
			name, _ := el.Attr("name")
			txt := el.Text()
			mp[name] = txt
		})
		Languages = append(Languages, "en")
		Words["en"] = mp
		fmt.Println("|")
	}
}

func Init() {
	raymond.RegisterHelper("translate", func(context interface{}, options *raymond.Options) string {
		a := options.Value("languages").(string)
		b := strings.Split(a, ",")
		id := options.ParamStr(0)
		b = append(b, "en")
		for _, item := range b {
			s, ok := GetWord(item, id)
			if ok {
				return s
			}
		}
		util.LogError("translations:", "error:", "unable to find value for:", id)
		return ""
	})
}

func GetWord(lang, word string) (string, bool) {
	if !stringsu.Contains(Languages, lang) {
		return "", false
	}
	w, ok := Words[lang][word]
	return w, ok
}
