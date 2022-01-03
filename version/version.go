package version

import (
	"fmt"
)

const (
	Version = "0.54"
)

// Used for "User-Agent" in HTTP
var VersionString = fmt.Sprintf("Go-WebRender %s", Version)

// reference commit 5ce71e48fe2deb745bb919876bcced91740316ba
