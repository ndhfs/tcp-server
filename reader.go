package tcp

import (
	"io"
)

type Reader interface {
	Read(reader io.Reader) ([]byte, error)
}
