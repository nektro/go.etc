package htp

// Controller is the htp package's extension
type Controller struct {
	//
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
