package tcp

type Handler func(c Context, m Msg) error
type ErrorHandler func(c Context, err error)

type EventType int
const (
	EventTypeAccepted EventType = iota+1
	EventTypeDisconnected
)
type Subscriber func(e EventType, c Context) error
