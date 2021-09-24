package tcp

type Handler interface {
	MessageReceived(ctx Conn, msg Msg) error
	ConnectionClosed(ctx Conn) error
}

type ErrorHandler func(ctx Conn, err error)
