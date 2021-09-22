package server

import (
	"context"
	"net"
)

type Listener interface {
	Listen(ctx context.Context, network string, addr string) (net.Listener, error)
}

type Option func(options *Options)

func NewSimpleListener() Listener {
	return new(net.ListenConfig)
}
