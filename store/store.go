package store

import (
	"sync"

	"github.com/nektro/go-util/util"

	. "github.com/nektro/go-util/alias"
)

// global singleton
var (
	This *Store
)

// Init takes flag values and initializes datastore
func Init(typ string, urlS string) {
	defer ensureStore()

	doInit := func() Inner {
		switch typ {
		}
		util.DieOnError(E(F("'%s' is not a valid store handler type", typ)))
		return nil
	}
	This = &Store{doInit()}
}

func ensureStore() {
	util.Log("store:", This.Type())
	util.DieOnError(This.Ping())
}

// Inner holds KV values
type Inner interface {
	Type() string
	Ping() error
	Has(key string) bool
	Set(key string, val string)
	Get(key string) string
	Range(f func(key string, val string) bool)
	HasList(key string) bool
	ListHas(key, value string) bool
	ListAdd(key, value string)
	ListRemove(key, value string)
	ListLen(key string) int
	ListGet(key string) []string
	sync.Locker
}

// Store is
type Store struct {
	Inner
}

// Keys returns array of all Keys
func (p Store) Keys() []string {
	res := []string{}
	p.Range(func(k, _ string) bool {
		res = append(res, k)
		return true
	})
	return res
}
