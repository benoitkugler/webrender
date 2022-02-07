package version

import (
	"fmt"
)

const (
	Version = "0.54"
)

// Used for "User-Agent" in HTTP
var VersionString = fmt.Sprintf("Go-WebRender %s", Version)

// commit of the Python reference implementation 4ddcd9504148374318ef91517c4144c84e5ba7e7
