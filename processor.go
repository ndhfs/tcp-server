package tcp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
)

type Processor interface {
	Listener
	Reader
	Writer
}

type RowSocketProcessor struct {
	readBufferSize int
}

func (r *RowSocketProcessor) Listen(ctx context.Context, network string, addr string) (net.Listener, error) {
	return new(net.ListenConfig).Listen(ctx, network, addr)
}

func NewRowSocketProcessor() *RowSocketProcessor {
	return &RowSocketProcessor{}
}

func (r *RowSocketProcessor) Read(reader io.Reader) ([]byte, error) {
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

func (r *RowSocketProcessor) Write(writer io.Writer, bytes []byte) error {
	_, err := writer.Write(append(bytes, '\n'))
	return err
}
