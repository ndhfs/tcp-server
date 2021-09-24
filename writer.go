package tcp

import (
	"io"
)

type Writer interface {
	Write(io.Writer, []byte) error
}

