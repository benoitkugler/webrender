package version

import (
	"fmt"
)

const (
	Version = "0.50"
)

// Used for "User-Agent" in HTTP
var VersionString = fmt.Sprintf("Go-WebRender %s", Version)
