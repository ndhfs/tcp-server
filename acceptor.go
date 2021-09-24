package tcp

import "net"

type Acceptor interface {
	Accept() (net.Conn, error)
}
