package tcp

import (
	"context"
	"net"
)

type Listener interface {
	Listen(ctx context.Context, network string, addr string) (net.Listener, error)
}
