package utils

import (
	"fmt"
)

const (
	Version = "0.62"
)

// Used for "User-Agent" in HTTP
var VersionString = fmt.Sprintf("Go-WebRender %s", Version)

// commit of the Python reference implementation d5d7ce369aef035712cf73446f9085a32105846f
