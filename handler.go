package tcp

type Handler func(c Conn, m Msg) error
type ErrorHandler func(c Conn, err error)

type EventType int
const (
	EventTypeAccepted EventType = iota+1
	EventTypeDisconnected
)
type Subscriber func(e EventType, c Conn) error
