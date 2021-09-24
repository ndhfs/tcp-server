package websocket

import (
	"fmt"
	"github.com/gobwas/ws"
	"io"
)

type Writer struct {
	opCode ws.OpCode
}

func (w *Writer) Write(wr io.Writer, b []byte) error {
	frame := ws.NewFrame(w.opCode, true, b)
	if err := ws.WriteFrame(wr, frame); err != nil {
		return fmt.Errorf("failed write frame. %w", err)
	}
	return nil
}

func NewWriter(opCode ws.OpCode) *Writer {
	return &Writer{opCode: opCode}
}

