package utils

import (
	"fmt"
)

const (
	Version = "0.59"
)

// Used for "User-Agent" in HTTP
var VersionString = fmt.Sprintf("Go-WebRender %s", Version)

// commit of the Python reference implementation 0ff8692741a58269e6b7e871819d018994579e16
