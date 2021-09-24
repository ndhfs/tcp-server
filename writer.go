package tcp

import (
	"bytes"
	"io"
)

type Writer interface {
	Write(io.Writer, []byte) error
}

type LineWriter struct {

}

func NewLineWriter() *LineWriter {
	return &LineWriter{}
}

func (l *LineWriter) Write(w io.Writer, b []byte) error {
	buf := bytes.NewBuffer(b)
	buf.WriteByte('\n')
	_, err := io.Copy(w, buf)
	return err
}

