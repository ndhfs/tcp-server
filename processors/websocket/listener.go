package websocket

import (
	"bufio"
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

	var isWs = true

	conn.SetReadDeadline(time.Now().Add(w.p.opts.backoffTimeout))
	br := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	first3Bytes, err := br.Peek(3)
	if err != nil {
		if erop, ok := err.(*net.OpError); ok && erop.Timeout() {
			isWs = false
		} else {
			conn.Close()
			return nil, fmt.Errorf("failed upgrate ws connection. %w", err)
		}
	} else if string(first3Bytes) == "GET" {
		_, err = w.u.Upgrade(br)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed upgrate ws connection. %w", err)
		}
		br.Flush()
	} else {
		isWs = false
	}

	return &Conn{Conn: conn, bf: br, isWs: isWs}, nil
}
