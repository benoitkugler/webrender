package logger

import (
	"log"
	"os"
)

var ProgressLogger = log.New(os.Stdout, "webrender.progress ", log.LstdFlags)
