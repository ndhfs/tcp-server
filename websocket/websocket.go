package websocket

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/ndhfs/tcp-server"
	"net"
)

func Opts(opCode ws.OpCode) []tcp.Option {
	return []tcp.Option{
		tcp.WithListener(NewListener()),
		tcp.WithReader(NewReader()),
		tcp.WithWriter(NewWriter(opCode)),
	}
}

type ListenConfig struct {
}

func NewListener() *ListenConfig {
	return &ListenConfig{}
}

func (l *ListenConfig) Listen(ctx context.Context, network string, addr string) (net.Listener, error) {
	nl, err := new(net.ListenConfig).Listen(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	return &Listener{
		Listener: nl,
		u:  ws.Upgrader{},
	}, nil
}
