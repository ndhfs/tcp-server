package tcp

type Middleware func(handler Handler) Handler
