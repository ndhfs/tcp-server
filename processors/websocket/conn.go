package websocket

import (
	"bufio"
	"net"
)

type Conn struct {
	net.Conn
	bf *bufio.ReadWriter
	isWs bool
}

func (b *Conn) Read(p []byte) (n int, err error) {
	return b.bf.Read(p)
}

func (b *Conn) Write(p []byte) (n int, err error) {
	n, err = b.bf.Write(p)
	if err != nil {
		return
	}
	return n, b.bf.Flush()
}
