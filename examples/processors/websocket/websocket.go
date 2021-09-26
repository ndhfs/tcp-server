package websocket

import (
	"context"
	"fmt"
	"github.com/gobwas/ws"
	"io"
	"net"
)

type Processor struct {
	opts Options
}

func NewProcessor(opts ...Option) *Processor {
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}
	return &Processor{opts: o}
}

func (p *Processor) Listen(ctx context.Context, network string, addr string) (net.Listener, error) {
	nl, err := new(net.ListenConfig).Listen(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	return &Listener{
		p:        p,
		Listener: nl,
		u:        ws.Upgrader{},
	}, nil
}

func (p *Processor) Read(in io.Reader) ([]byte, error) {
	conn := in.(*Conn)
	if !conn.isWs && p.opts.backoffProcessor != nil {
		return p.opts.backoffProcessor.Read(in)
	}

	hdr, err := ws.ReadHeader(in)
	if err != nil {
		return nil, fmt.Errorf("failed read websocket header. %w", err)
	}

	if hdr.OpCode == ws.OpClose {
		return nil, io.EOF
	}

	payload := make([]byte, hdr.Length)
	_, err = io.ReadFull(in, payload)
	if err != nil {
		// handle error
	}
	if hdr.Masked {
		ws.Cipher(payload, hdr.Mask, 0)
	}
	return payload, err
}

func (p *Processor) Write(wr io.Writer, b []byte) error {
	conn := wr.(*Conn)
	if !conn.isWs && p.opts.backoffProcessor != nil {
		return p.opts.backoffProcessor.Write(wr, b)
	}

	frame := ws.NewFrame(p.opts.opCode, true, b)
	if err := ws.WriteFrame(wr, frame); err != nil {
		return fmt.Errorf("failed write frame. %p", err)
	}
	return nil
}
