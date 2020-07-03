package internal

import (
	"time"
)

// globals used in sub packages but dont need cgo
var (
	Epoch, _ = time.Parse("Jan 2 2006", "Jan 1 2020")
)
