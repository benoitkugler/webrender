// Package logger provides two log.Logger emitting progress status and warning
// information.
package logger

import (
	"log"
	"os"
)

// ProgressLogger logs the main steps of the HTLM rendering.
// It is purely informative and may be turned off safely.
var ProgressLogger = log.New(os.Stdout, "webrender.progress: ", log.LstdFlags)

// WarningLogger emits a warning for each non fatal error, like unsupported CSS
// properties, font loading or URL resolutions errors.
// It can be turned off safely, but it is a good source of information if the
// rendering seems wrong.
var WarningLogger = log.New(os.Stdout, "webrender.warning: ", log.Lmsgprefix)
