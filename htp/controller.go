package htp

import (
	"fmt"
	"net/http"
	"strconv"
)

// Controller is the htp package's extension
type Controller struct {
	r *http.Request
}

// panic tests condition and panics if not met
func (v *Controller) panic(t string, b bool, d string) {
	if !b {
		panic("htp: " + t + ": " + d)
	}
}

// Assert will exit this http method if condition is not met
func (v *Controller) Assert(condition bool, message string) {
	v.panic("assertion failed", condition, message)
}

// RedirectIf will redirct to location if condition is met
func (v *Controller) RedirectIf(condition bool, location string) {
	v.panic("redirect", !condition, location)
}

func (v *Controller) GetQueryString(name string) string {
	s := v.r.URL.Query().Get(name)
	v.Assert(len(s) > 0, "400: missing query value: "+name)
	return s
}

func (v *Controller) GetFormString(name string) string {
	s := v.r.Form.Get(name)
	v.Assert(len(s) > 0, "400: missing form value: "+name)
	return s
}

func (v *Controller) GetFormInt(name string) (string, int64) {
	s := v.GetFormString(name)
	n, err := strconv.ParseInt(s, 10, 64)
	v.Assert(err == nil, "400: form value must be a number: "+name)
	return s, n
}

func (v *Controller) AssertNilErr(err error) {
	v.Assert(err == nil, "400: "+fmt.Sprintf("%v", err))
}
