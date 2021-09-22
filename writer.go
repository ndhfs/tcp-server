package server

import (
	"io"
)

type Writer interface {
	Write(writer io.Writer, data []byte) error
}

type LineWriter struct {

}

func NewLineWriter() *LineWriter {
	return &LineWriter{}
}

func (l *LineWriter) Write(writer io.Writer, data []byte) error {
	_, err := writer.Write(append(data, '\n'))
	return err
}

