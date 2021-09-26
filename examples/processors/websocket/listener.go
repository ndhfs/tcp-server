package websocket

import (
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"net"
	"time"
)

type Listener struct {
	net.Listener
	p *Processor
	u ws.Upgrader
}

func (w *Listener) Accept() (net.Conn, error) {
	conn, err := w.Listener.Accept()
	if err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(w.p.opts.backoffTimeout))
	_, err = w.u.Upgrade(conn)
	var isWs = true
	if err != nil {
		if w.p.opts.backoffProcessor != nil {
			if errors.Is(err, ws.ErrMalformedRequest) {
				isWs = false
			} else if erop, ok := err.(*net.OpError); ok && erop.Timeout() {
				isWs = false
			} else {
				conn.Close()
				return nil, fmt.Errorf("failed upgrate ws connection. %w", err)
			}
		} else {
			conn.Close()
			return nil, fmt.Errorf("failed upgrate ws connection. %w", err)
		}
	}

	return &Conn{Conn: conn, isWs: isWs}, nil
}
