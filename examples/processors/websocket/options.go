package websocket

import (
	"github.com/gobwas/ws"
	"github.com/ndhfs/tcp-server"
	"time"
)

type Options struct {
	opCode           ws.OpCode
	backoffProcessor tcp.Processor
	backoffTimeout   time.Duration
}

type Option func(options *Options)

var defaultOptions = Options{
	opCode:         ws.OpText,
	backoffTimeout: 1 * time.Millisecond,
}

func WithOpCode(code ws.OpCode) Option {
	return func(options *Options) {
		options.opCode = code
	}
}

func WithBackoffProcessor(processor tcp.Processor) Option {
	return func(options *Options) {
		options.backoffProcessor = processor
	}
}

func WithBackoffTimeout(duration time.Duration) Option {
	return func(options *Options) {
		options.backoffTimeout = duration
	}
}
