package tcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type Conn interface {
	ID() string
	RemoteAddr() net.Addr
	LocalAddr() net.Addr

	Duration() time.Duration

	Send(v Msg) error
	Close() error

	Context() context.Context
}

type connection struct {
	id        string
	conn      net.Conn
	createdAt time.Time
	doneCtx   context.Context
	doneFn    func()
	closed    bool
	mu        sync.Mutex
	s         *Server
}

func (c *connection) Context() context.Context {
	return c.doneCtx
}

func (c *connection) ID() string {
	return c.id
}

func (c *connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *connection) Duration() time.Duration {
	return time.Now().Sub(c.createdAt)
}

func (c *connection) Send(v Msg) error {
	var err error

	if enc := c.s.opts.encoder; enc != nil {
		v, err = enc.Encode(v)
		if err != nil {
			return fmt.Errorf("failed encode message. %w", err)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.s.opts.writeTimeout > 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.s.opts.writeTimeout))
	}

	if bb, ok := v.([]byte); ok {
		return c.s.opts.processor.Write(c.conn, bb)
	}

	return fmt.Errorf("final Msg must be []byte")
}

func (c *connection) Close() error {
	c.mu.Lock()
	if c.closed {
		return nil
	}
	c.closed = true
	c.mu.Unlock()
	c.doneFn()
	c.conn.Close()
	c.s.handleEvent(EventTypeDisconnected, c)
	return nil
}

