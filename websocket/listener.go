package websocket

import (
	"fmt"
	"github.com/gobwas/ws"
	"net"
)

type Listener struct {
	net.Listener
	u  ws.Upgrader
}

func (w *Listener) Accept() (net.Conn, error) {
	conn, err := w.Listener.Accept()
	if err != nil {
		return nil, err
	}

	_, err = w.u.Upgrade(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed upgrate ws connection. %w", err)
	}

	return conn, err
}
