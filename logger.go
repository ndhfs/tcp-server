package server

import (
	"log"
	"os"
)

type Logger interface {
	// Printf
	// Arguments are handled in the manner of fmt.Printf.
	Printf(format string, v ...interface{})
}

func NewStdLogger() Logger {
	return log.New(os.Stdout, "", log.LstdFlags)
}
