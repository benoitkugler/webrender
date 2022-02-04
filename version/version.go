package version

import (
	"fmt"
)

const (
	Version = "0.54"
)

// Used for "User-Agent" in HTTP
var VersionString = fmt.Sprintf("Go-WebRender %s", Version)

// TODO: update to last version
// commit of the Python reference implementation 117bbae615bc6b01c93cbab937c387e06dd1ae4e
