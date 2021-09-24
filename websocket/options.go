package websocket

import (
	"github.com/gobwas/ws"
	"github.com/ndhfs/tcp-server"
)

type Options struct {
	opCode ws.OpCode
	alternativeProcessor tcp.Processor
}

type Option func(options *Options)

var defaultOptions = Options{
	opCode: ws.OpText,
}

func WithOpCode(code ws.OpCode) Option {
	return func(options *Options) {
		options.opCode = code
	}
}

func WithAlternativeProcessor(processor tcp.Processor) Option {
	return func(options *Options) {
		options.alternativeProcessor = processor
	}
}
