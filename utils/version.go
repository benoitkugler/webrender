package utils

import (
	"fmt"
)

const (
	Version = "0.59"
)

// Used for "User-Agent" in HTTP
var VersionString = fmt.Sprintf("Go-WebRender %s", Version)

// commit of the Python reference implementation 2a9a952c356b856ad1f80ddb6d0c93b36ccac46c
