package tcp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type Reader interface {
	Read(reader io.Reader) ([]byte, error)
}

type LineReader struct {
	bufferSize uint
}

func NewLineReader(bufferSize uint) *LineReader {
	return &LineReader{bufferSize: bufferSize}
}

func (f *LineReader) Read(reader io.Reader) ([]byte, error) {
	scanner := bufio.NewReader(reader)
	var buf bytes.Buffer
	for {
		line, isPrefix, err := scanner.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("LineReader. failed read stream. %w", err)
		}
		buf.Write(line)
		if !isPrefix {
			break
		}
	}
	return buf.Bytes(), nil
}

