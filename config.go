package etc

import (
	"encoding/json"
	"reflect"

	. "github.com/nektro/go-util/alias"
	. "github.com/nektro/go-util/util"
)

func InitConfig(path string, template interface{}) {
	DieOnError(Assert(DoesFileExist(path), path+" does not exist!"))
	json.Unmarshal(ReadFile(path), &template)
}

func ConfigAssertKeysNonEmpty(config interface{}, keys ...string) {
	v := reflect.ValueOf(config).Elem().Elem()
	for _, k := range keys {
		f := v.FieldByName(k).String()
		t, _ := v.Type().FieldByName(k)
		g := t.Tag.Get("json")
		DieOnError(Assert(f != "", F("config[%s] is empty!", g)))
	}
}
