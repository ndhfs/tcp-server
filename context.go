package tcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type Context interface {
	ID() string
	RemoteAddr() net.Addr
	LocalAddr() net.Addr

	Duration() time.Duration

	Send(v Msg) error
	Close() error
	CloseWithErr(err error) error

	Context() context.Context

	Set(key string, v interface{})
	Get(key string) (v interface{})
}

type Map map[string]interface{}

type connContext struct {
	mu        sync.Mutex
	id        string
	conn      net.Conn
	createdAt time.Time
	doneCtx   context.Context
	doneFn    func()
	closed    bool
	s         *Server
	store     Map
}

func (c *connContext) CloseWithErr(err error) error {
	c.s.handleError(c, err)
	return c.Close()
}

func (c *connContext) Set(key string, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.store == nil {
		c.store = make(Map)
	}

	c.store[key] = v
}

func (c *connContext) Get(key string) (v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.store[key]
}

func (c *connContext) Context() context.Context {
	return c.doneCtx
}

func (c *connContext) ID() string {
	return c.id
}

func (c *connContext) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *connContext) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *connContext) Duration() time.Duration {
	return time.Now().Sub(c.createdAt)
}

func (c *connContext) Send(v Msg) error {
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

func (c *connContext) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()
	c.doneFn()
	c.conn.Close()
	c.s.handleEvent(EventTypeDisconnected, c)
	return nil
}

