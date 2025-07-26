package utils

import (
	"fmt"
)

const (
	Version = "0.62"
)

// Used for "User-Agent" in HTTP
var VersionString = fmt.Sprintf("Go-WebRender %s", Version)

// commit of the Python reference implementation 5f7c4e6ae1ec9c65d978ea61333512fc62425c4b
