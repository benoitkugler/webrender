package logger

import (
	"log"
	"os"
)

// ProgressLogger logs the main steps of the HTLM rendering.
var ProgressLogger = log.New(os.Stdout, "webrender.progress: ", log.LstdFlags)

// WarningLogger emits a warning for each non fatal error, like unsupported CSS
// properties, font loading errors or URL resolutions
var WarningLogger = log.New(os.Stdout, "webrender.warning: ", log.Lmsgprefix)
