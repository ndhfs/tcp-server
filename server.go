package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	opts   Options
	l      net.Listener
	runCtx context.Context
	doneFn func()
	wg     sync.WaitGroup
	mu     sync.RWMutex
	conns  map[string]*connection
	ws     chan struct{}
	ath    chan struct{}

	handler Handler
	errorHandler ErrorHandler
}

func (s *Server) SetErrorHandler(errorHandler ErrorHandler) {
	s.errorHandler = errorHandler
}

func (s *Server) Serve(ctx context.Context, network string, addr string) error {
	s.logInfo("Start serving at %s %s", network, addr)
	defer func() {
		s.logInfo("Stop serving at %s %s", network, addr)
	}()
	var err error
	s.l, err = s.opts.listener.Listen(ctx, network, addr)
	if err != nil {
		return fmt.Errorf("failed start listener. %w", err)
	}

	for {

		select {
		case s.ath <- struct{}{}:
		default:
			time.Sleep(1 * time.Millisecond)
			continue
		}

		conn, err := s.l.Accept()
		if err != nil {
			select {
			case <-s.runCtx.Done():
				return nil
			default:
				if operr, ok := err.(*net.OpError); ok {
					if operr.Temporary() {
						s.logDebug("failed accept conn. %s", err)
						<-s.ath
						continue
					}
				}
				return fmt.Errorf("failed accept conn. %w", err)
			}
		}

		c := s.newConnection(conn)
		go s.handleConnection(c)
		<-s.ath
	}
}

func (s *Server) newConnection(conn net.Conn) *connection {
	c := &connection{
		id:        uuid.New().String(),
		conn:      conn,
		createdAt: time.Now(),
		s:         s,
	}

	c.doneCtx, c.doneFn = context.WithCancel(s.runCtx)

	s.mu.Lock()
	s.conns[c.id] = c
	s.mu.Unlock()
	return c
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logInfo("Shutting down tcp server")
	defer func() {
		s.logInfo("Shutting down tcp server complete")
	}()

	// signal all connections for shutting down
	s.doneFn()

	// stop accept new conns
	s.l.Close()

	return nil
}

func (s *Server) GracefulShutdown(ctx context.Context) error {
	s.logInfo("Graceful shutting down tcp server")
	defer func() {
		s.logInfo("Graceful shutting down tcp server complete")
	}()

	// signal all connections for shutting down
	s.doneFn()

	// stop accept new conns
	s.l.Close()

	var waitCh = make(chan struct{})
	go func() {
		s.logInfo("waiting active connections closed. Num cons: %d", s.NumConns())
		s.wg.Wait()
		close(waitCh)
	}()

	for {
		select {
		case <-waitCh:
			s.logInfo("waiting active connections closed. Success")
			return nil
		case <-time.Tick(5 * time.Second):
			s.logInfo("waiting active connections closed. Num cons: %d", s.NumConns())
		case <-ctx.Done():
			s.logErr("waiting active connections closed. Timout expired. Force closed")
			return nil
		}
	}
}

func (s *Server) handleConnection(c *connection) {
	s.wg.Add(1)
	s.logDebug("New connection %s", c.conn.RemoteAddr())
	defer func() {
		c.Close()
		s.mu.Lock()
		delete(s.conns, c.id)
		s.mu.Unlock()

		s.logDebug("Connection %s closed.", c.conn.RemoteAddr())
		s.wg.Done()
	}()

	for {
		// Если завершились, больше не слушать
		select {
		case <-c.doneCtx.Done():
			return
		default:
		}

		c.conn.SetReadDeadline(time.Now().Add(s.opts.readTimeout))
		b, err := s.opts.reader.Read(c.conn)

		// Если завершились, не обрабатываем
		select {
		case <-c.doneCtx.Done():
			return
		default:
		}

		if err != nil {
			var erop *net.OpError
			if errors.Is(err, io.EOF) {
			} else if errors.As(err, &erop) && erop.Timeout() {
			} else {
				s.logErr("Err read conn. %s", err)
			}
			return
		}

		if err := s.dispatch(c, b); err != nil {
			s.handleError(c, err)
		}
	}
}

func New(opts ...Option) *Server {
	var o = defaultOptions
	for _, opt := range opts {
		opt(&o)
	}

	ctx, doneFn := context.WithCancel(context.Background())
	return &Server{
		opts:   o,
		runCtx: ctx,
		doneFn: doneFn,
		conns:  make(map[string]*connection, 100),
		ws:     make(chan struct{}, o.workersNum),
		ath:    make(chan struct{}, o.acceptThreshold),
	}
}

func (s *Server) SetHandler(handler Handler) {
	s.handler = handler
}

func (s *Server) logErr(format string, v ...interface{}) {
	s.opts.logger.Printf("[ERR] "+format, v...)
}

func (s *Server) logDebug(format string, v ...interface{}) {
	if s.opts.debug {
		s.opts.logger.Printf("[DBG] "+format, v...)
	}
}

// logInfo
func (s *Server) logInfo(format string, a ...interface{}) {
	s.opts.logger.Printf("[INF] "+format, a...)
}

func (s *Server) NumConns() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.conns)
}

func (s *Server) dispatch(c *connection, b []byte) error {
	select {
	case s.ws <- struct{}{}:
		return s.dispatchAsync(c, b)
	case <-time.After(s.opts.workerWaitTimeout):
		return ErrServerIsBusy
	}
}

func (s *Server) handleError(c *connection, err error) {
	if s.errorHandler != nil {
		s.errorHandler(c, err)
	} else {
		s.logErr("failed dispatch client message. %s", err)
	}
}

func (s *Server) dispatchAsync(c *connection, b []byte) error {
	go func() {
		defer func() {
			<-s.ws
		}()

		// process message
		s.logDebug("Received msg: %s", string(b))
		if s.handler == nil {
			s.logInfo("Handler not registered. Skipping")
			return
		}

		msg, err := s.opts.encoder.Decode(b)
		if err != nil {
			s.handleError(c, err)
			return
		}

		if err := s.handler.MessageReceived(c, msg); err != nil {
			s.handleError(c, err)
			return
		}
	}()
	return nil
}
