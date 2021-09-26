package websocket

import "net"

type Conn struct {
	net.Conn
	isWs bool
}
